package ai

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/jpf"
)

func NewFileQATool(builder *ModelBuilder) agent.Tool {
	return &batchFileQA{builder}
}

type fileQAInput struct {
	Content string
	Query   string
}

type batchFileQA struct {
	builder *ModelBuilder
}

func (b *batchFileQA) Call(args map[string]any) (string, error) {
	mf := b.buildMF()
	queryAny, ok1 := args["query"]
	pathsAny, ok2 := args["paths"]
	if !(ok1 && ok2) {
		return "", errors.New("must provide a query and paths field")
	}
	query, ok1 := queryAny.(string)
	pathsAnyList, ok2 := pathsAny.([]any)
	if !(ok1 && ok2) {
		return "", errors.New("query must be a string and paths muct be a list of strings")
	}
	paths := make([]string, len(pathsAnyList))
	for i := range paths {
		paths[i], ok1 = pathsAnyList[i].(string)
		if !ok1 {
			return "", errors.New("all paths must be strings")
		}
	}
	contents := make([]string, len(paths))
	for i, pth := range paths {
		err := func() error {
			f, err := os.Open(pth)
			if err != nil {
				return err
			}
			defer f.Close()
			data, err := io.ReadAll(f)
			if err != nil {
				return err
			}
			contents[i] = string(data)
			return nil
		}()
		if err != nil {
			return "", err
		}
	}
	results := make([]string, len(paths))
	errs := make([]error, len(paths))
	wg := &sync.WaitGroup{}
	wg.Add(len(paths))
	for i, content := range contents {
		go func() {
			defer wg.Done()
			res, _, err := mf.Call(fileQAInput{content, query})
			if err != nil {
				errs[i] = err
			} else {
				results[i] = fmt.Sprintf("%s:\n%s", paths[i], res)
			}
		}()
	}
	wg.Wait()
	errs = slices.DeleteFunc(errs, func(err error) bool { return err == nil })
	if len(errs) != 0 {
		return "", errors.Join(errs...)
	}
	return strings.Join(results, "\n\n"), nil
}

func (b *batchFileQA) buildMF() jpf.MapFunc[fileQAInput, string] {
	return jpf.NewOneShotMapFunc(
		jpf.NewTemplateMessageEncoder[fileQAInput]("", "Document:\n{{ .Content }}\n\n\nQuery: {{ .Query }}"),
		jpf.NewRawStringResponseDecoder(),
		b.builder.BuildFileQAModel(),
	)
}

func (b *batchFileQA) Description() []string {
	return []string{
		"Run a plaintext query across one or more files.",
		"The query can be as simple or advanced as you need.",
		"Pass two arguments, 'query' (string) and 'paths' (list of strings).",
	}
}

func (b *batchFileQA) Name() string {
	return "query_files"
}
