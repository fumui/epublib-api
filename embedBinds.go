package epublib

import "embed"

//go:embed docs/swagger.yaml
var SwaggerSpec []byte

//go:embed docs/swaggerui
var SwaggerUI embed.FS
