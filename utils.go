package agent

import (
	"bytes"
	"net/url"
	"strconv"
	"text/template"
)

func parseUrlParamsToArgs(encoded string) (map[string]any, error) {
	values, err := url.ParseQuery(encoded)
	if err != nil {
		return nil, err
	}

	out := make(map[string]any)
	for k, v := range values {
		if len(v) == 0 {
			continue
		}
		s := v[0]

		if b, err := strconv.ParseBool(s); err == nil {
			out[k] = b
		} else if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			out[k] = i
		} else if f, err := strconv.ParseFloat(s, 64); err == nil {
			out[k] = f
		} else {
			out[k] = s
		}
	}

	return out, nil
}

func formatTemplate(tpl string, data any) (string, error) {
	tplCompiled, err := template.New("prompt").Parse(tpl)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	err = tplCompiled.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
