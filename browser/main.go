package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"browser/css"
	"browser/dom"
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
	browser.OnNavigate = func(newURL string) {
		loadPage(browser, newURL)
	}

	// Load initial page
	loadPage(browser, startURL)

	// Run the GUI
	browser.Run()
}

func loadPage(browser *render.Browser, pageURL string) {
	fmt.Println("Fetching:", pageURL)
	browser.ShowLoading()
	browser.UpdateURLBar(pageURL)

	// Run fetch in background so UI stays responsive
	go func() {
		resp, err := http.Get(pageURL)
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
