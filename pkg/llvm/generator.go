package llvm

import (
	"embed"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/utils"
)

//go:embed templates
var Templates embed.FS

type Generator struct {
	template *template.Template
}

func NewGenerator() (*Generator, error) {
	funcs :=
		template.FuncMap{
			"ToUpper": strings.ToUpper,
			"ToLower": strings.ToLower,
			"String":  fmt.Sprint,
			"Binary": func(bits int, value uint64) string {
				return "0b" + utils.FormatUintBinary(value, bits)
			},
			"Join": func(separator string, items []string) string {
				return strings.Join(items, separator)
			},
			"LLVMType": instructions.LLVMType,
			"MapMember": func(member string, items any) ([]any, error) {
				v := reflect.ValueOf(items)
				if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
					return nil, fmt.Errorf("expected array, got %v", v.Kind())
				}

				arr := make([]any, v.Len())

				for i := 0; i < v.Len(); i++ {
					arr[i] = v.Index(i).Interface()
				}

				return utils.MapMember(member, arr)
			},
		}

	funcs["MapStrings"] =
		func(funcName string, items []any) ([]string, error) {
			var function any
			hasFunction := false
			if function, hasFunction = funcs[funcName]; !hasFunction {
				return nil, fmt.Errorf("function '%v' not found in template funcs", function)
			}

			f := function.(func(string) string)

			if f == nil {
				return nil, fmt.Errorf("function '%v' does not gave signature func(string) string", funcName)
			}

			return utils.Map(items, func(str any) string {
				return f(str.(string))
			}), nil
		}

	t, err := template.New("Cucaracha.td").Funcs(funcs).
		ParseFS(Templates, "templates/Cucaracha*.td")

	if err != nil {
		return nil, err
	}

	return &Generator{
		template: t,
	}, nil
}

func (g *Generator) GenerateTo(writer io.Writer) error {
	return g.template.Execute(writer, &mc.Descriptor)
}

func (g *Generator) Generate(outputFile string) error {
	f, err := os.Create(outputFile)

	if err != nil {
		return err
	}

	return g.GenerateTo(f)
}
