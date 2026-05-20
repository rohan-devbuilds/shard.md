package tools

import "strings"

type Request struct {
	Action string
	Arg    string
}

func ParseToolRequest(text string) (Request, bool) {
	start := strings.Index(text, "```tool_request")
	if start < 0 {
		if req, ok := parseToolRequestBody(text); ok {
			return req, true
		}
		return InferToolRequest(text)
	}
	rest := text[start+len("```tool_request"):]
	end := strings.Index(rest, "```")
	if end < 0 {
		return Request{}, false
	}
	return parseToolRequestBody(rest[:end])
}

func InferToolRequest(text string) (Request, bool) {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "list the files") || strings.Contains(lower, "list files") || strings.Contains(lower, "inspect the repo") || strings.Contains(lower, "inspect the repository") || strings.Contains(lower, "look at the repo") || strings.Contains(lower, "look through the files"):
		return Request{Action: "list_dir", Arg: "."}, true
	case strings.Contains(lower, "run the tests") || strings.Contains(lower, "run tests") || strings.Contains(lower, "go test ./...") || strings.Contains(lower, "check tests"):
		return Request{Action: "run_command", Arg: "go test ./..."}, true
	case strings.Contains(lower, "build the project") || strings.Contains(lower, "go build .") || strings.Contains(lower, "compile the project"):
		return Request{Action: "run_command", Arg: "go build ."}, true
	case strings.Contains(lower, "read the project") || strings.Contains(lower, "read all files") || strings.Contains(lower, "analyze the project files") || strings.Contains(lower, "read the files"):
		return Request{Action: "read_project", Arg: "."}, true
	}
	return Request{}, false
}

func parseToolRequestBody(body string) (Request, bool) {
	body = strings.TrimSpace(body)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "read_file", "read_project", "list_dir", "run_command":
			if value != "" {
				return Request{Action: key, Arg: value}, true
			}
		}
	}
	return Request{}, false
}
