package hcj

func DefaultCSS() CSS {
	return CSS{
		Selectors: map[string]map[string]string{
			"a:link": {
				"color":           "blue",
				"text-decoration": "underline",
				"cursor":          "auto",
			},
			"a:visited": {
				"color":           "dark-blue",
				"text-decoration": "underline",
				"cursor":          "auto",
			},
			"a:link:active": {
				"color": "light-blue",
			},
			"a:visited:active": {
				"color": "dark-blue",
			},
			"address": {
				"display":    "block",
				"font-style": "italic",
			},
			"area": {
				"display": "none",
			},
			"article": {
				"display": "block",
			},
			"aside": {
				"display": "block",
			},
			"b": {
				"font-weight": "bold",
			},
			"bdo": {
				"unicode-bidi": "bidi-override",
			},
			"blockquote": {
				"display":       "block",
				"margin-top":    "1em",
				"margin-bottom": "1em",
				"margin-left":   "40px",
				"margin-right":  "40px",
			},
			"body": {
				"display": "block",
				"margin":  "8px",
			},
			"body:focus": {
				"outline": "none",
			},
			"caption": {
				"display":    "table-caption",
				"text-align": "center",
			},
			"cite": {
				"font-style": "italic",
			},
			"code": {
				"font-family": "monospace",
			},
			"col": {
				"display": "table-column",
			},
			"colgroup": {
				"display": "table-column-group",
			},
			"datalist": {
				"display": "none",
			},
			"dd": {
				"display":     "block",
				"margin-left": "40px",
			},
			"del": {
				"text-decoration": "line-through",
			},
			"details": {
				"display": "block",
			},
			"dfn": {
				"font-style": "italic",
			},
			"div": {
				"display": "block",
			},
			"dl": {
				"display":       "block",
				"margin-top":    "1em",
				"margin-bottom": "1em",
				"margin-left":   "0",
				"margin-right":  "0",
			},
			"dt": {
				"display": "block",
			},
			"em": {
				"font-style": "italic",
			},
			"embed:focus": {
				"outline": "none",
			},
			"fieldset": {
				"display":        "block",
				"margin-left":    "2px",
				"margin-right":   "2px",
				"padding-top":    "0.35em",
				"padding-bottom": "0.625em",
				"padding-left":   "0.75em",
				"padding-right":  "0.75em",
				"border":         "2px",
			},
			"figcaption": {
				"display": "block",
			},
			"figure": {
				"display":       "block",
				"margin-top":    "1em",
				"margin-bottom": "1em",
				"margin-left":   "40px",
				"margin-right":  "40px",
			},
			"footer": {
				"display": "block",
			},
			"form": {
				"display":    "block",
				"margin-top": "0em",
			},
			"h1": {
				"display":       "block",
				"font-size":     "2em",
				"margin-top":    "0.67em",
				"margin-bottom": "0.67em",
				"margin-left":   "0",
				"margin-right":  "0",
				"font-weight":   "bold",
			},
			"h2": {
				"display":       "block",
				"font-size":     "1.5em",
				"margin-top":    "0.83em",
				"margin-bottom": "0.83em",
				"margin-left":   "0",
				"margin-right":  "0",
				"font-weight":   "bold",
			},
			"h3": {
				"display":       "block",
				"font-size":     "1.17em",
				"margin-top":    "1em",
				"margin-bottom": "1em",
				"margin-left":   "0",
				"margin-right":  "0",
				"font-weight":   "bold",
			},
			"h4": {
				"display":       "block",
				"margin-top":    "1.33em",
				"margin-bottom": "1.33em",
				"margin-left":   "0",
				"margin-right":  "0",
				"font-weight":   "bold",
			},
			"h5": {
				"display":       "block",
				"font-size":     ".83em",
				"margin-top":    "1.67em",
				"margin-bottom": "1.67em",
				"margin-left":   "0",
				"margin-right":  "0",
				"font-weight":   "bold",
			},
			"h6": {
				"display":       "block",
				"font-size":     ".67em",
				"margin-top":    "2.33em",
				"margin-bottom": "2.33em",
				"margin-left":   "0",
				"margin-right":  "0",
				"font-weight":   "bold",
			},
			"head": {
				"display": "none",
			},
			"header": {
				"display": "block",
			},
			"hr": {
				"display":       "block",
				"margin-top":    "0.5em",
				"margin-bottom": "0.5em",
				"margin-left":   "auto",
				"margin-right":  "auto",
				"border-style":  "inset",
				"border-width":  "1px",
			},
			"html": {
				"display": "block",
			},
			"html:focus": {
				"outline": "none",
			},
			"i": {
				"font-style": "italic",
			},
			"iframe:focus": {
				"outline": "none",
			},
			"iframe[seamless]": {
				"display": "block",
			},
			"img": {
				"display": "inline-block",
			},
			"ins": {
				"text-decoration": "underline",
			},
			"kbd": {
				"font-family": "monospace",
			},
			"label": {
				"cursor": "default",
			},
			"legend": {
				"display":       "block",
				"padding-left":  "2px",
				"padding-right": "2px",
				"border":        "none",
			},
			"li": {
				"display": "list-item",
			},
			"link": {
				"display": "none",
			},
			"map": {
				"display": "inline",
			},
			"mark": {
				"background-color": "yellow",
				"color":            "black",
			},
			"menu": {
				"display":         "block",
				"list-style-type": "disc",
				"margin-top":      "1em",
				"margin-bottom":   "1em",
				"margin-left":     "0",
				"margin-right":    "0",
				"padding-left":    "40px",
			},
			"nav": {
				"display": "block",
			},
			"object:focus": {
				"outline": "none",
			},
			"ol": {
				"display":         "block",
				"list-style-type": "decimal",
				"margin-top":      "1em",
				"margin-bottom":   "1em",
				"margin-left":     "0",
				"margin-right":    "0",
				"padding-left":    "40px",
			},
			"output": {
				"display": "inline",
			},
			"p": {
				"display":       "block",
				"margin-top":    "1em",
				"margin-bottom": "1em",
				"margin-left":   "0",
				"margin-right":  "0",
			},
			"param": {
				"display": "none",
			},
			"pre": {
				"display":     "block",
				"font-family": "monospace",
				"white-space": "pre",
				"margin":      "1em 0",
			},
			"q": {
				"display": "inline",
			},
			"q::before": {
				"content": "open-quote",
			},
			"q::after": {
				"content": "close-quote",
			},
			"rt": {
				"line-height": "normal",
			},
			"s": {
				"text-decoration": "line-through",
			},
			"samp": {
				"font-family": "monospace",
			},
			"script": {
				"display": "none",
			},
			"section": {
				"display": "block",
			},
			"small": {
				"font-size": "smaller",
			},
			"strike": {
				"text-decoration": "line-through",
			},
			"strong": {
				"font-weight": "bold",
			},
			"style": {
				"display": "none",
			},
			"sub": {
				"vertical-align": "sub",
				"font-size":      "smaller",
			},
			"summary": {
				"display": "block",
			},
			"sup": {
				"vertical-align": "super",
				"font-size":      "smaller",
			},
			"table": {
				"display":         "table",
				"border-collapse": "separate",
				"border-spacing":  "2px",
				"border-color":    "gray",
			},
			"tbody": {
				"display":        "table-row-group",
				"vertical-align": "middle",
				"border-color":   "inherit",
			},
			"td": {
				"display":        "table-cell",
				"vertical-align": "inherit",
			},
			"tfoot": {
				"display":        "table-footer-group",
				"vertical-align": "middle",
				"border-color":   "inherit",
			},
			"th": {
				"display":        "table-cell",
				"vertical-align": "inherit",
				"font-weight":    "bold",
				"text-align":     "center",
			},
			"thead": {
				"display":        "table-header-group",
				"vertical-align": "middle",
				"border-color":   "inherit",
			},
			"title": {
				"display": "none",
			},
			"tr": {
				"display":        "table-row",
				"vertical-align": "inherit",
				"border-color":   "inherit",
			},
			"u": {
				"text-decoration": "underline",
			},
			"ul": {
				"display":         "block",
				"list-style-type": "disc",
				"margin-top":      "1em",
				"margin-bottom":   "1 em",
				"margin-left":     "0",
				"margin-right":    "0",
				"padding-left":    "40px",
			},
			"var": {
				"font-style": "italic",
			},
		},
	}
}
