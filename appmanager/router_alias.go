package appmanager

import "github.com/ozanturksever/uiwgo/router"

// RouteDefinitionAlias is an alias to router.RouteDefinition to avoid importing router in types.go
// while keeping a light dependency surface in the types file.
type RouteDefinitionAlias = router.RouteDefinition
