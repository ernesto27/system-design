package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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
	browser := render.NewBrowser(800, 600)

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
		var fullCSS string

		// 1. Fetch external stylesheets
		links := dom.FindStylesheetLinks(document)
		for _, link := range links {
			absURL := resolveURL(pageURL, link)
			fmt.Println("Fetching CSS:", absURL)
			cssResp, err := http.Get(absURL)
			if err == nil {
				data, _ := io.ReadAll(cssResp.Body)
				fullCSS += string(data) + "\n"
				cssResp.Body.Close()
			} else {
				fmt.Println("Failed to fetch CSS:", err)
			}
		}

		// 2. Add internal <style> content
		fullCSS += dom.FindStyleContent(document)

		fmt.Println("Building layout...")
		stylesheet := css.Parse(fullCSS)
		browser.SetStylesheet(stylesheet)
		layoutTree := layout.BuildLayoutTree(document, stylesheet)
		layout.ComputeLayout(layoutTree, 800)

		// Execute JavaScript
		fmt.Println("Executing JavaScript...")
		jsRuntime := js.NewJSRuntime(document, func() {
			fmt.Println("DOM changed, would reflow here")
		})

		browser.SetJSClickHandler(jsRuntime.DispatchClick)

		scripts := js.FindScripts(document)
		for i, script := range scripts {
			fmt.Printf("Running script %d...\n", i+1)
			jsRuntime.Execute(script)
		}

		browser.SetCurrentURL(pageURL)
		browser.SetContent(layoutTree)
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
