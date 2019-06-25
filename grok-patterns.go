package main

import "github.com/vjeantet/grok"

func AddDefaultPatterns(g *grok.Grok) {
	// Nginx
	g.AddPattern("NGINX_ERROR_DATESTAMP", `\d{4}/\d{2}/\d{2}[- ]%{TIME}`)
	g.AddPattern("NGINX_ERROR_LOG", `%{NGINX_ERROR_DATESTAMP:timestamp} \[%{DATA:err_severity}\] (%{NUMBER:pid:int}#%{NUMBER}: \*%{NUMBER}|\*%{NUMBER}) %{DATA:message}(?:, client: "?%{IPORHOST:client}"?)(?:, server: %{IPORHOST:server})(?:, request: "%{WORD:verb} %{URIPATHPARAM:request} HTTP/%{NUMBER:httpversion}")?(?:, upstream: "%{DATA:upstream}")?(?:, host: "%{URIHOST:host}")?(?:, referrer: "%{URI:referrer}")?`)
}
