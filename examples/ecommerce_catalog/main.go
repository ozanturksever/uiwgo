//go:build js && wasm

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ozanturksever/uiwgo/comps"
	"github.com/ozanturksever/uiwgo/dom"
	"github.com/ozanturksever/uiwgo/reactivity"

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
		products:         reactivity.CreateSignal([]Product{}),
		searchTerm:       reactivity.CreateSignal(""),
		selectedCategory: reactivity.CreateSignal(""),
		viewMode:         reactivity.CreateSignal(ViewModeGrid),
		sortBy:           reactivity.CreateSignal(SortByName),
		sortAsc:          reactivity.CreateSignal(true),
		showOutOfStock:   reactivity.CreateSignal(true),
		loading:          reactivity.CreateSignal(false),
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
	// Load sample data on mount
	comps.OnMount(func() {
		pc.loadSampleData()
	})

	// Computed filtered and sorted products
	filteredProducts := reactivity.CreateMemo(func() []Product {
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
	categories := reactivity.CreateMemo(func() []string {
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

	return h.Div(
		h.Class("product-catalog"),

		// Header with filters and controls
		h.Div(
			h.Class("catalog-header"),
			h.Style("padding: 2rem; background: #f8f9fa; margin-bottom: 2rem;"),
			h.H1(g.Text("Product Catalog")),

			// Search bar
			h.Div(
				h.Class("search-bar"),
				h.Style("margin: 1rem 0;"),
				h.Input(
					h.ID("search-input"),
					h.Type("text"),
					h.Placeholder("Search products..."),
					h.Value(pc.searchTerm.Get()),
					h.Style("padding: 0.5rem; width: 300px; border: 1px solid #ddd; border-radius: 4px;"),
				),
				comps.OnMount(func() {
					if searchInput := dom.GetElementByID("search-input"); searchInput != nil {
						dom.BindInputToSignal(searchInput, pc.searchTerm)
					}
				}),
			),

			// Filters row
			h.Div(
				h.Class("filters-row"),
				h.Style("display: flex; gap: 1rem; align-items: center; flex-wrap: wrap;"),

				// Category filter
			h.Div(
					h.Label(
						h.Style("margin-right: 0.5rem;"),
						g.Text("Category: "),
					),
					h.Select(
						h.ID("category-select"),
						h.Value(pc.selectedCategory.Get()),
						h.Style("padding: 0.25rem;"),
						h.Option(h.Value(""), g.Text("All Categories")),
						comps.For(comps.ForProps[string]{
							Items: categories,
							Key:   func(cat string) string { return cat },
							Children: func(cat string, index int) g.Node {
								return h.Option(
									h.Value(cat),
									g.Text(cat),
								)
							},
						}),
					),
					comps.OnMount(func() {
						if categorySelect := dom.GetElementByID("category-select"); categorySelect != nil {
							dom.BindChangeToSignal(categorySelect, pc.selectedCategory)
						}
					}),
			),

		// Sort controls
		h.Div(
					h.Label(
						h.Style("margin-right: 0.5rem;"),
						g.Text("Sort by: "),
					),
					h.Select(
						h.ID("sort-select"),
						h.Value(string(pc.sortBy.Get())),
						h.Style("padding: 0.25rem;"),
						h.Option(h.Value(string(SortByName)), g.Text("Name")),
						h.Option(h.Value(string(SortByPrice)), g.Text("Price")),
						h.Option(h.Value(string(SortByRating)), g.Text("Rating")),
						h.Option(h.Value(string(SortByCategory)), g.Text("Category")),
					),
					comps.OnMount(func() {
					if sortSelect := dom.GetElementByID("sort-select"); sortSelect != nil {
						dom.BindChange(sortSelect, func(event dom.Event) {
							if target := event.Target(); target != nil {
								value := target.Underlying().Get("value").String()
								pc.sortBy.Set(SortBy(value))
							}
						})
					}
				}),
				),

				h.Button(
					h.ID("sort-direction-btn"),
					h.Style("padding: 0.25rem 0.5rem; margin-left: 0.5rem;"),
					comps.BindText(func() string {
						if pc.sortAsc.Get() {
							return "↑ Asc"
						}
						return "↓ Desc"
					}),
				),
				comps.OnMount(func() {
					if sortBtn := dom.GetElementByID("sort-direction-btn"); sortBtn != nil {
						dom.BindClickToCallback(sortBtn, func() {
							pc.sortAsc.Set(!pc.sortAsc.Get())
						})
					}
				}),

				// View mode toggle
				h.Div(
					h.Class("view-toggle"),
					h.Style("margin-left: 1rem;"),
					h.Button(
						h.ID("grid-view-btn"),
						h.Style(func() string {
					style := "padding: 0.25rem 0.5rem; margin-right: 0.25rem;"
					if pc.viewMode.Get() == ViewModeGrid {
						style += " background: #007bff; color: white;"
					}
					return style
				}()),
						g.Text("Grid"),
					),
					h.Button(
						h.ID("list-view-btn"),
						h.Style(func() string {
					style := "padding: 0.25rem 0.5rem;"
					if pc.viewMode.Get() == ViewModeList {
						style += " background: #007bff; color: white;"
					}
					return style
				}()),
						g.Text("List"),
					),
					comps.OnMount(func() {
						if gridBtn := dom.GetElementByID("grid-view-btn"); gridBtn != nil {
							dom.BindClickToCallback(gridBtn, func() {
								pc.viewMode.Set(ViewModeGrid)
							})
						}
						if listBtn := dom.GetElementByID("list-view-btn"); listBtn != nil {
							dom.BindClickToCallback(listBtn, func() {
								pc.viewMode.Set(ViewModeList)
							})
						}
					}),
				),

				// Show out of stock toggle
				h.Label(
					h.Style("margin-left: 1rem;"),
					h.Input(
				h.ID("show-out-of-stock-checkbox"),
				h.Type("checkbox"),
				g.If(pc.showOutOfStock.Get(), h.Checked()),
			),
					g.Text(" Show out of stock"),
				),
				comps.OnMount(func() {
					if checkbox := dom.GetElementByID("show-out-of-stock-checkbox"); checkbox != nil {
						dom.BindChange(checkbox, func(event dom.Event) {
							if target := event.Target(); target != nil {
								checked := target.Underlying().Get("checked").Bool()
								pc.showOutOfStock.Set(checked)
							}
						})
					}
				}),
			),
		),

		// Loading state
		comps.Show(comps.ShowProps{
			When: pc.loading,
			Children: h.Div(
				h.Class("loading-state"),
				h.Style("text-align: center; padding: 2rem; font-size: 1.2rem;"),
				g.Text("Loading products..."),
			),
		}),

		// Products display
		comps.Show(comps.ShowProps{
			When: reactivity.CreateMemo(func() bool {
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
			When: reactivity.CreateMemo(func() bool {
				return !pc.loading.Get() && len(filteredProducts.Get()) == 0
			}),
			Children: h.Div(
				h.Class("empty-state"),
				h.Style("text-align: center; padding: 3rem; color: #666;"),
				h.H3(g.Text("No products found")),
				h.P(g.Text("Try adjusting your filters or search terms.")),
			),
		}),
	)
}

func (pc *ProductCatalog) renderProductGrid(products reactivity.Signal[[]Product]) g.Node {
	return h.Div(
		h.Class("product-grid"),
		h.Style("display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 1rem; padding: 1rem;"),
		comps.For(comps.ForProps[Product]{
			Items: products,
			Key:   func(p Product) string { return p.ID },
			Children: func(p Product, index int) g.Node {
				return h.Div(
					h.Class("product-card"),
					h.Style(func() string {
						style := "border: 1px solid #ddd; border-radius: 8px; padding: 1rem; background: white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
						if !p.InStock {
							style += " opacity: 0.6;"
						}
						return style
					}()),

					h.Img(
						h.Src(p.ImageURL),
						h.Alt(p.Name),
						h.Class("product-image"),
						h.Style("width: 100%; height: 200px; object-fit: cover; border-radius: 4px; margin-bottom: 1rem;"),
					),

					h.Div(
						h.Class("product-info"),
						h.H3(
							h.Style("margin: 0 0 0.5rem 0; font-size: 1.1rem;"),
							g.Text(p.Name),
						),
						h.P(
							h.Class("category"),
							h.Style("color: #666; font-size: 0.9rem; margin: 0 0 0.5rem 0;"),
							g.Text(p.Category),
						),
						h.P(
							h.Class("price"),
							h.Style("font-size: 1.2rem; font-weight: bold; color: #007bff; margin: 0 0 0.5rem 0;"),
							g.Text(fmt.Sprintf("$%.2f", p.Price)),
						),
						h.Div(
							h.Class("rating"),
							h.Style("color: #ffa500; margin-bottom: 0.5rem;"),
							g.Text(fmt.Sprintf("★ %.1f", p.Rating)),
						),
						comps.Show(comps.ShowProps{
							When: reactivity.CreateSignal(!p.InStock),
							Children: h.Span(
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
	return h.Div(
		h.Class("product-list"),
		h.Style("padding: 2rem;"),
		comps.For(comps.ForProps[Product]{
			Items: products,
			Key:   func(p Product) string { return p.ID },
			Children: func(p Product, index int) g.Node {
				return h.Div(
					h.Class("product-row"),
					h.Style(func() string {
						style := "display: flex; gap: 1rem; padding: 1rem; border: 1px solid #ddd; border-radius: 8px; margin-bottom: 1rem; background: white;"
						if !p.InStock {
							style += " opacity: 0.6;"
						}
						return style
					}()),

					h.Img(
						h.Src(p.ImageURL),
						h.Alt(p.Name),
						h.Class("product-thumbnail"),
						h.Style("width: 100px; height: 100px; object-fit: cover; border-radius: 4px; flex-shrink: 0;"),
					),

					h.Div(
						h.Class("product-details"),
						h.Style("flex: 1;"),
						h.H3(
							h.Style("margin: 0 0 0.5rem 0; font-size: 1.2rem;"),
							g.Text(p.Name),
						),
						h.P(
							h.Class("description"),
							h.Style("color: #666; margin: 0 0 1rem 0; line-height: 1.4;"),
							g.Text(p.Description),
						),
						h.Div(
							h.Class("product-meta"),
							h.Style("display: flex; gap: 1rem; align-items: center;"),
							h.Span(
								h.Class("category"),
								h.Style("background: #f8f9fa; padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.9rem;"),
								g.Text(p.Category),
							),
							h.Span(
								h.Class("price"),
								h.Style("font-size: 1.3rem; font-weight: bold; color: #007bff;"),
								g.Text(fmt.Sprintf("$%.2f", p.Price)),
							),
							h.Span(
								h.Class("rating"),
								h.Style("color: #ffa500;"),
								g.Text(fmt.Sprintf("★ %.1f", p.Rating)),
							),
							comps.Show(comps.ShowProps{
							When: reactivity.CreateSignal(!p.InStock),
								Children: h.Span(
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
	// Mount the app and get a disposer function
	disposer := comps.Mount("app", func() g.Node {
		return NewProductCatalog().render()
	})
	_ = disposer // We don't use it in this example since the app runs indefinitely

	// Prevent exit
	select {}
}
