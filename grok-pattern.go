package main

import "github.com/vjeantet/grok"

func AddDefaultPatterns(g *grok.Grok) {
	g.AddPattern("NGINX_ERROR_LOG", `%{DATESTAMP:timestamp} \[%{DATA:err_severity}\] (%{NUMBER:pid:int}#%{NUMBER}: \*%{NUMBER}|\*%{NUMBER}) %{DATA:err_message}(?:, client: "?%{IPORHOST:client}"?)(?:, server: %{IPORHOST:server})(?:, request: "%{WORD:verb} %{URIPATHPARAM:request} HTTP/%{NUMBER:httpversion}")?(?:, upstream: "%{DATA:upstream}")?(?:, host: "%{URIHOST:host}")?(?:, referrer: "%{URI:referrer}")?`)
}
