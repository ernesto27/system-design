package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"browser/css"
	"browser/dom"
	"browser/js"
	"browser/layout"
	"browser/render"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <url>")
		os.Exit(1)
	}

	startURL := os.Args[1]

	// Create browser window
	browser := render.NewBrowser(900, 600)

	// When link is clicked or Go pressed, load the page
	browser.OnNavigate = func(req render.NavigationRequest) {
		loadPage(browser, req)
	}
	// Load initial page
	loadPage(browser, render.NavigationRequest{
		URL:    startURL,
		Method: "GET",
	})

	// Run the GUI
	browser.Run()
}

func loadPage(browser *render.Browser, req render.NavigationRequest) {
	pageURL := req.URL
	method := req.Method
	if method == "" {
		method = "GET"
	}
	fmt.Printf("Fetching (%s): %s\n", method, pageURL)
	browser.ShowLoading()
	browser.UpdateURLBar(pageURL)

	// Run fetch in background so UI stays responsive
	go func() {
		var resp *http.Response
		var err error

		if method == "POST" {
			if req.Body != nil && req.ContentType != "" {
				// Multipart form data (file upload)
				httpReq, err := http.NewRequest("POST", pageURL, bytes.NewReader(req.Body))
				if err != nil {
					fmt.Println("Error creating request:", err)
					browser.ShowError("Error creating request")
					return
				}
				httpReq.Header.Set("Content-Type", req.ContentType)
				resp, err = http.DefaultClient.Do(httpReq)
			} else {
				// URL-encoded form data (default)
				resp, err = http.PostForm(pageURL, req.Data)
			}
		} else {
			resp, err = http.Get(pageURL)
		}

		if err != nil {
			fmt.Println("Error:", err)
			browser.ShowError("Error 404")
			return
		}
		defer resp.Body.Close()

		fmt.Println("Parsing HTML...")
		document := dom.Parse(resp.Body)
		if document == nil {
			browser.ShowError("Error 404")
			fmt.Println("Error: failed to parse HTML")
			return
		}

		title := dom.FindTitle(document)
		browser.SetTitle(title)
		browser.SetDocument(document)

		fmt.Println("Fetching CSS...")

		// 1. Fetch external stylesheets in parallel
		links := dom.FindStylesheetLinks(document)
		cssResults := make([]string, len(links))
		var wg sync.WaitGroup

		for i, link := range links {
			wg.Add(1)
			go func(idx int, href string) {
				defer wg.Done()
				absURL := resolveURL(pageURL, href)
				fmt.Println("Fetching CSS:", absURL)
				cssResp, err := http.Get(absURL)
				if err == nil {
					data, _ := io.ReadAll(cssResp.Body)
					cssResults[idx] = string(data)
					cssResp.Body.Close()
				} else {
					fmt.Println("Failed to fetch CSS:", err)
				}
			}(i, link)
		}

		wg.Wait()

		// Combine external CSS in order
		var externalCSS strings.Builder
		for _, cssContent := range cssResults {
			externalCSS.WriteString(cssContent + "\n")
		}

		// Store external CSS for reflow (when styles are disabled/enabled)
		browser.SetExternalCSS(externalCSS.String())

		// Combine external + internal <style> content
		fullCSS := externalCSS.String() + dom.FindActiveStyleContent(document)

		fmt.Println("Building layout...")
		stylesheet := css.Parse(fullCSS)
		browser.SetDocument(document)
		layoutTree := layout.BuildLayoutTree(document, stylesheet, layout.Viewport{
			Width:  float64(browser.Width),
			Height: float64(browser.Height),
		})
		layout.ComputeLayout(layoutTree, float64(browser.Width))

		// Execute JavaScript
		fmt.Println("Executing JavaScript...")
		jsRuntime := js.NewJSRuntime(document, func() {
			browser.Reflow(browser.Width)
		})

		jsRuntime.SetAlertHandler(browser.ShowAlert)
		jsRuntime.SetConfirmHandler(browser.ShowConfirm)
		jsRuntime.SetPromptHandler(browser.ShowPrompt)
		browser.SetJSClickHandler(jsRuntime.DispatchClick)
		browser.SetBeforeNavigateHandler(jsRuntime.CheckBeforeUnload)

		jsRuntime.SetCurrentURL(pageURL)

		scripts := js.FindScripts(document)
		for i, script := range scripts {
			fmt.Printf("Running script %d...\n", i+1)
			jsRuntime.Execute(script)
		}

		browser.SetCurrentURL(pageURL)
		jsRuntime.SetReloadHandler(func() {
			browser.Refresh()
		})

		jsRuntime.SetTitleChangeHandler(browser.SetTitle)

		// Re-parse CSS after JavaScript (respects disabled styles)
		fullCSS = externalCSS.String() + dom.FindActiveStyleContent(document)
		stylesheet = css.Parse(fullCSS)

		// Rebuild layout tree AFTER JavaScript has modified the DOM
		layoutTree = layout.BuildLayoutTree(document, stylesheet, layout.Viewport{
			Width:  float64(browser.Width),
			Height: float64(browser.Height),
		})
		layout.ComputeLayout(layoutTree, float64(browser.Width))
		browser.SetContent(layoutTree)

		bodyNode := dom.FindElementsByTagName(document, dom.TagBody)
		if bodyNode != nil {
			if onload, ok := bodyNode.Attributes["onload"]; ok {
				fmt.Println("Executing body onload...")
				jsRuntime.Execute(onload)
			}
		}

		browser.AddToHistory(pageURL)

		fmt.Println("Page loaded!")
	}()
}

func resolveURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}
