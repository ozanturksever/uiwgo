//go:build js && wasm

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"
	"github.com/ozanturksever/uiwgo/wasm"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type Product struct {
	ID          string
	Name        string
	Price       float64
	Category    string
	ImageURL    string
	Description string
	InStock     bool
	Rating      float64
}

type ViewMode string

const (
	ViewModeGrid ViewMode = "grid"
	ViewModeList ViewMode = "list"
)

type SortBy string

const (
	SortByName     SortBy = "name"
	SortByPrice    SortBy = "price"
	SortByRating   SortBy = "rating"
	SortByCategory SortBy = "category"
)

type ProductCatalog struct {
	products         reactivity.Signal[[]Product]
	searchTerm       reactivity.Signal[string]
	selectedCategory reactivity.Signal[string]
	viewMode         reactivity.Signal[ViewMode]
	sortBy           reactivity.Signal[SortBy]
	sortAsc          reactivity.Signal[bool]
	showOutOfStock   reactivity.Signal[bool]
	loading          reactivity.Signal[bool]
}

func NewProductCatalog() *ProductCatalog {
	return &ProductCatalog{
		products:         reactivity.NewSignal([]Product{}),
		searchTerm:       reactivity.NewSignal(""),
		selectedCategory: reactivity.NewSignal(""),
		viewMode:         reactivity.NewSignal(ViewModeGrid),
		sortBy:           reactivity.NewSignal(SortByName),
		sortAsc:          reactivity.NewSignal(true),
		showOutOfStock:   reactivity.NewSignal(true),
		loading:          reactivity.NewSignal(false),
	}
}

func (pc *ProductCatalog) loadSampleData() {
	// In a real app, this would be an API call
	sampleProducts := []Product{
		{ID: "1", Name: "Wireless Headphones", Price: 99.99, Category: "Electronics", ImageURL: "https://via.placeholder.com/200x200?text=Headphones", Description: "High-quality wireless headphones with noise cancellation", InStock: true, Rating: 4.5},
		{ID: "2", Name: "Smartphone", Price: 699.99, Category: "Electronics", ImageURL: "https://via.placeholder.com/200x200?text=Phone", Description: "Latest smartphone with advanced camera", InStock: true, Rating: 4.8},
		{ID: "3", Name: "Running Shoes", Price: 129.99, Category: "Sports", ImageURL: "https://via.placeholder.com/200x200?text=Shoes", Description: "Comfortable running shoes for all terrains", InStock: false, Rating: 4.2},
		{ID: "4", Name: "Coffee Maker", Price: 79.99, Category: "Home", ImageURL: "https://via.placeholder.com/200x200?text=Coffee", Description: "Automatic coffee maker with timer", InStock: true, Rating: 4.0},
		{ID: "5", Name: "Laptop", Price: 1299.99, Category: "Electronics", ImageURL: "https://via.placeholder.com/200x200?text=Laptop", Description: "High-performance laptop for work and gaming", InStock: true, Rating: 4.7},
		{ID: "6", Name: "Yoga Mat", Price: 29.99, Category: "Sports", ImageURL: "https://via.placeholder.com/200x200?text=Yoga", Description: "Non-slip yoga mat for all exercises", InStock: true, Rating: 4.3},
		{ID: "7", Name: "Desk Lamp", Price: 49.99, Category: "Home", ImageURL: "https://via.placeholder.com/200x200?text=Lamp", Description: "Adjustable LED desk lamp", InStock: false, Rating: 4.1},
		{ID: "8", Name: "Bluetooth Speaker", Price: 59.99, Category: "Electronics", ImageURL: "https://via.placeholder.com/200x200?text=Speaker", Description: "Portable Bluetooth speaker with great sound", InStock: true, Rating: 4.4},
	}
	
	pc.products.Set(sampleProducts)
}

func (pc *ProductCatalog) render() g.Node {
	// Computed filtered and sorted products
	filteredProducts := reactivity.NewMemo(func() []Product {
		products := pc.products.Get()
		search := strings.ToLower(pc.searchTerm.Get())
		category := pc.selectedCategory.Get()
		showOOS := pc.showOutOfStock.Get()

		var filtered []Product
		for _, p := range products {
			// Category filter
			if category != "" && p.Category != category {
				continue
			}

			// Search filter
			if search != "" {
				if !strings.Contains(strings.ToLower(p.Name), search) &&
					!strings.Contains(strings.ToLower(p.Description), search) {
					continue
				}
			}

			// Stock filter
			if !showOOS && !p.InStock {
				continue
			}

			filtered = append(filtered, p)
		}

		// Sort products
		sortBy := pc.sortBy.Get()
		asc := pc.sortAsc.Get()

		sort.Slice(filtered, func(i, j int) bool {
			var less bool
			switch sortBy {
			case SortByName:
				less = filtered[i].Name < filtered[j].Name
			case SortByPrice:
				less = filtered[i].Price < filtered[j].Price
			case SortByRating:
				less = filtered[i].Rating < filtered[j].Rating
			case SortByCategory:
				less = filtered[i].Category < filtered[j].Category
			}

			if asc {
				return less
			}
			return !less
		})

		return filtered
	})

	// Get unique categories
	categories := reactivity.NewMemo(func() []string {
		products := pc.products.Get()
		catMap := make(map[string]bool)
		for _, p := range products {
			catMap[p.Category] = true
		}

		var cats []string
		for cat := range catMap {
			cats = append(cats, cat)
		}
		sort.Strings(cats)
		return cats
	})

	return g.Div(
		h.Class("product-catalog"),

		// Header with filters and controls
		g.Div(
			h.Class("catalog-header"),
			h.Style("padding: 2rem; background: #f8f9fa; margin-bottom: 2rem;"),
			g.H1(g.Text("Product Catalog")),

			// Search bar
			g.Div(
				h.Class("search-bar"),
				h.Style("margin: 1rem 0;"),
				g.Input(
					h.Type("text"),
					h.Placeholder("Search products..."),
					h.Value(pc.searchTerm.Get()),
					h.Style("padding: 0.5rem; width: 300px; border: 1px solid #ddd; border-radius: 4px;"),
					dom.OnInput(func(value string) {
						pc.searchTerm.Set(value)
					}),
				),
			),

			// Filters row
			g.Div(
				h.Class("filters-row"),
				h.Style("display: flex; gap: 1rem; align-items: center; flex-wrap: wrap;"),

				// Category filter
				g.Div(
					g.Label(
						h.Style("margin-right: 0.5rem;"),
						g.Text("Category: "),
					),
					g.Select(
						h.Value(pc.selectedCategory.Get()),
						h.Style("padding: 0.25rem;"),
						dom.OnChange(func(value string) {
							pc.selectedCategory.Set(value)
						}),
						g.Option(h.Value(""), g.Text("All Categories")),
						comps.For(comps.ForProps[string]{
							Items: categories,
							Key:   func(cat string) string { return cat },
							Children: func(cat string, index int) g.Node {
								return g.Option(
									h.Value(cat),
									g.Text(cat),
								)
							},
						}),
					),
				),

				// Sort controls
				g.Div(
					g.Label(
						h.Style("margin-right: 0.5rem;"),
						g.Text("Sort by: "),
					),
					g.Select(
						h.Value(string(pc.sortBy.Get())),
						h.Style("padding: 0.25rem;"),
						dom.OnChange(func(value string) {
							pc.sortBy.Set(SortBy(value))
						}),
						g.Option(h.Value(string(SortByName)), g.Text("Name")),
						g.Option(h.Value(string(SortByPrice)), g.Text("Price")),
						g.Option(h.Value(string(SortByRating)), g.Text("Rating")),
						g.Option(h.Value(string(SortByCategory)), g.Text("Category")),
					),
				),

				g.Button(
					h.Style("padding: 0.25rem 0.5rem; margin-left: 0.5rem;"),
					g.Text(func() string {
						if pc.sortAsc.Get() {
							return "↑ Asc"
						}
						return "↓ Desc"
					}()),
					dom.OnClick(func() {
						pc.sortAsc.Set(!pc.sortAsc.Get())
					}),
				),

				// View mode toggle
				g.Div(
					h.Class("view-toggle"),
					h.Style("margin-left: 1rem;"),
					g.Button(
						h.Style(func() string {
							style := "padding: 0.25rem 0.5rem; margin-right: 0.25rem;"
							if pc.viewMode.Get() == ViewModeGrid {
								style += " background: #007bff; color: white;"
							}
							return style
						}()),
						g.Text("Grid"),
						dom.OnClick(func() {
							pc.viewMode.Set(ViewModeGrid)
						}),
					),
					g.Button(
						h.Style(func() string {
							style := "padding: 0.25rem 0.5rem;"
							if pc.viewMode.Get() == ViewModeList {
								style += " background: #007bff; color: white;"
							}
							return style
						}()),
						g.Text("List"),
						dom.OnClick(func() {
							pc.viewMode.Set(ViewModeList)
						}),
					),
				),

				// Show out of stock toggle
				g.Label(
					h.Style("margin-left: 1rem;"),
					g.Input(
						h.Type("checkbox"),
						h.Checked(pc.showOutOfStock.Get()),
						dom.OnChange(func() {
							pc.showOutOfStock.Set(!pc.showOutOfStock.Get())
						}),
					),
					g.Text(" Show out of stock"),
				),
			),
		),

		// Loading state
		comps.Show(comps.ShowProps{
			When: pc.loading,
			Children: g.Div(
				h.Class("loading-state"),
				h.Style("text-align: center; padding: 2rem; font-size: 1.2rem;"),
				g.Text("Loading products..."),
			),
		}),

		// Products display
		comps.Show(comps.ShowProps{
			When: reactivity.NewMemo(func() bool {
				return !pc.loading.Get()
			}),
			Children: comps.Switch(comps.SwitchProps{
				When: pc.viewMode,
				Children: []g.Node{
					comps.Match(comps.MatchProps{
						When:     ViewModeGrid,
						Children: pc.renderProductGrid(filteredProducts),
					}),
					comps.Match(comps.MatchProps{
						When:     ViewModeList,
						Children: pc.renderProductList(filteredProducts),
					}),
				},
			}),
		}),

		// Empty state
		comps.Show(comps.ShowProps{
			When: reactivity.NewMemo(func() bool {
				return !pc.loading.Get() && len(filteredProducts.Get()) == 0
			}),
			Children: g.Div(
				h.Class("empty-state"),
				h.Style("text-align: center; padding: 3rem; color: #666;"),
				g.H3(g.Text("No products found")),
				g.P(g.Text("Try adjusting your filters or search terms.")),
			),
		}),
	)
}

func (pc *ProductCatalog) renderProductGrid(products reactivity.Signal[[]Product]) g.Node {
	return g.Div(
		h.Class("product-grid"),
		h.Style("display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 1.5rem; padding: 2rem;"),
		comps.For(comps.ForProps[Product]{
			Items: products,
			Key:   func(p Product) string { return p.ID },
			Children: func(p Product, index int) g.Node {
				return g.Div(
					h.Class("product-card"),
					h.Style(func() string {
						style := "border: 1px solid #ddd; border-radius: 8px; padding: 1rem; background: white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
						if !p.InStock {
							style += " opacity: 0.6;"
						}
						return style
					}()),

					g.Img(
						h.Src(p.ImageURL),
						h.Alt(p.Name),
						h.Class("product-image"),
						h.Style("width: 100%; height: 200px; object-fit: cover; border-radius: 4px; margin-bottom: 1rem;"),
					),

					g.Div(
						h.Class("product-info"),
						g.H3(
							h.Style("margin: 0 0 0.5rem 0; font-size: 1.1rem;"),
							g.Text(p.Name),
						),
						g.P(
							h.Class("category"),
							h.Style("color: #666; font-size: 0.9rem; margin: 0 0 0.5rem 0;"),
							g.Text(p.Category),
						),
						g.P(
							h.Class("price"),
							h.Style("font-size: 1.2rem; font-weight: bold; color: #007bff; margin: 0 0 0.5rem 0;"),
							g.Text(fmt.Sprintf("$%.2f", p.Price)),
						),
						g.Div(
							h.Class("rating"),
							h.Style("color: #ffa500; margin-bottom: 0.5rem;"),
							g.Text(fmt.Sprintf("★ %.1f", p.Rating)),
						),
						comps.Show(comps.ShowProps{
							When: reactivity.NewSignal(!p.InStock),
							Children: g.Span(
								h.Class("stock-status"),
								h.Style("background: #dc3545; color: white; padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.8rem;"),
								g.Text("Out of Stock"),
							),
						}),
					),
				)
			},
		}),
	)
}

func (pc *ProductCatalog) renderProductList(products reactivity.Signal[[]Product]) g.Node {
	return g.Div(
		h.Class("product-list"),
		h.Style("padding: 2rem;"),
		comps.For(comps.ForProps[Product]{
			Items: products,
			Key:   func(p Product) string { return p.ID },
			Children: func(p Product, index int) g.Node {
				return g.Div(
					h.Class("product-row"),
					h.Style(func() string {
						style := "display: flex; gap: 1rem; padding: 1rem; border: 1px solid #ddd; border-radius: 8px; margin-bottom: 1rem; background: white;"
						if !p.InStock {
							style += " opacity: 0.6;"
						}
						return style
					}()),

					g.Img(
						h.Src(p.ImageURL),
						h.Alt(p.Name),
						h.Class("product-thumbnail"),
						h.Style("width: 100px; height: 100px; object-fit: cover; border-radius: 4px; flex-shrink: 0;"),
					),

					g.Div(
						h.Class("product-details"),
						h.Style("flex: 1;"),
						g.H3(
							h.Style("margin: 0 0 0.5rem 0; font-size: 1.2rem;"),
							g.Text(p.Name),
						),
						g.P(
							h.Class("description"),
							h.Style("color: #666; margin: 0 0 1rem 0; line-height: 1.4;"),
							g.Text(p.Description),
						),
						g.Div(
							h.Class("product-meta"),
							h.Style("display: flex; gap: 1rem; align-items: center;"),
							g.Span(
								h.Class("category"),
								h.Style("background: #f8f9fa; padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.9rem;"),
								g.Text(p.Category),
							),
							g.Span(
								h.Class("price"),
								h.Style("font-size: 1.3rem; font-weight: bold; color: #007bff;"),
								g.Text(fmt.Sprintf("$%.2f", p.Price)),
							),
							g.Span(
								h.Class("rating"),
								h.Style("color: #ffa500;"),
								g.Text(fmt.Sprintf("★ %.1f", p.Rating)),
							),
							comps.Show(comps.ShowProps{
								When: reactivity.NewSignal(!p.InStock),
								Children: g.Span(
									h.Class("stock-status"),
									h.Style("background: #dc3545; color: white; padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.8rem;"),
									g.Text("Out of Stock"),
								),
							}),
						),
					),
				)
			},
		}),
	)
}

func main() {
	wasm.Initialize()

	catalog := NewProductCatalog()

	// Load sample data on startup
	reactivity.NewEffect(func() {
		catalog.loadSampleData()
	})

	dom.Mount("#app", catalog.render())

	wasm.KeepAlive()
}