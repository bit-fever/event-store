//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package messaging

import (
	"os"

	"github.com/bit-fever/core"
	"gopkg.in/yaml.v3"
)

//=============================================================================

type Template struct {
	Code    string
	Level   string
	Title   string
	Message string
}

//=============================================================================

var templates = make(map[string]*Template)

//=============================================================================

func initTemplates() {
	var content map[string]interface{}

	file, err := os.ReadFile("config/event-templates.yaml")
	core.ExitIfError(err)

	err = yaml.Unmarshal(file, &content)
	core.ExitIfError(err)

	buildTemplateMap("", content)
}

//=============================================================================

func buildTemplateMap(path string, currMap map[string]interface{}) {
	title, ok  := currMap["title"]
	message, _ := currMap["message"]
	level ,  _ := currMap["level"]

	if ok {
		templates[path] = &Template{
			Code   : path,
			Level  : level.(string),
			Title  : title.(string),
			Message: message.(string),
		}
	} else {
		for key, value := range currMap {
			subPath := path + "." + key

			if path == "" {
				subPath = key
			}

			buildTemplateMap(subPath, value.(map[string]interface{}))
		}
	}
}

//=============================================================================
