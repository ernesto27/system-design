package main

import (
	"fmt"
	"net/http"
	"os"

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

func loadPage(browser *render.Browser, url string) {
	fmt.Println("Fetching:", url)
	browser.ShowLoading()
	browser.UpdateURLBar(url)

	// Run fetch in background so UI stays responsive
	go func() {
		resp, err := http.Get(url)
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

		fmt.Println("Building layout...")
		layoutTree := layout.BuildLayoutTree(document)
		layout.ComputeLayout(layoutTree, 800)

		browser.SetCurrentURL(url)
		browser.SetContent(layoutTree)
		browser.AddToHistory(url)

		fmt.Println("Page loaded!")
	}()
}
