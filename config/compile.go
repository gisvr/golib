package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/gisvr/golib/log"

	"github.com/flosch/pongo2"
)

var (
	noDotRegexp   = regexp.MustCompile("{{ *([^\\.][^ ]+) *}}")
	indentRegexp  = regexp.MustCompile("([ ]+){{ *indent +([^ }]+) *}} *")
	encryptRegexp = regexp.MustCompile("{{ *encrypt +([^\\.\"'][^\"' ]+) *}}")
	boolRegexp    = regexp.MustCompile(`'\s*{{\s*bool +([^ ]+)\s*}}\s*'`)
)

var funcMap = template.FuncMap{
	"indent":  indent,
	"encrypt": encrypt,
}

// SetFuncMap update the default funcmap
func SetFuncMap(k string, f interface{}) {
	funcMap[k] = f
}

// CompileT ...
type CompileT struct {
	Env  string
	vars map[string]string
}

// SetVar set kv into self
func (c *CompileT) SetVar(k, v string) {
	c.vars[k] = v
}

// SetVars set a lot of kv into self
func (c *CompileT) SetVars(vars map[string]string) {
	for k, v := range vars {
		c.vars[k] = v
	}
}

// ResetVars reset all vars
func (c *CompileT) ResetVars(vars map[string]string) {
	c.vars = make(map[string]string)
	c.SetVars(vars)
}

// NewCompileT returns a new pointer
func NewCompileT(env string) *CompileT {
	c := &CompileT{
		Env:  env,
		vars: make(map[string]string),
	}

	return c
}

func hasVariable(s string) bool {
	return strings.Index(s, "{{") != -1
}

func indent(in string, count int) string {
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return in
	}
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", count) + line
	}

	return strings.Join(lines, "\n")
}

func encrypt(in string) string {
	out, err := Encrypt(in)
	if err != nil {
		log.Fatalf("error encrypt: %v", err.Error())
	}
	return out
}

// LoadVarsFromString 从字符串中读取kv 值，并更新到r结构的var中
// 这个函数内部负责了各种tricky的事情，比如变量替换，env的关键字替换，函数操作（如encrypt)等
func LoadVarsFromString(str string, r *CompileT) error {
	dat := []byte(str)

	var localVars map[string]interface{}

	err := yaml.Unmarshal(dat, &localVars)
	if err != nil {
		log.Infof("error load vars from string\n%v", str)

		return err
	}

	var ssMap = make(map[string]*template.Template)

	for k, v := range localVars {
		if r.Env != "" && k[len(k)-1] == ']' {
			ks := strings.Split(k, ".[")
			if len(ks) == 2 {
				log.Infof("extract %v to %v", k, ks)

				k = ks[0]
				env := strings.TrimSuffix(ks[1], "]")
				if env != r.Env {
					if env == "default" {
						if _, ok := localVars[fmt.Sprintf("%s.[%v]", k, r.Env)]; ok {
							log.Infof("skip key %v, using %v", ks, r.Env)
							continue
						}

						log.Infof("using %v[default]", k)
					} else {
						log.Infof("skip key %v", ks)
						continue
					}
				}
			}
		}

		var sv string
		sv, err = extractString(k, v, r)
		if err != nil {
			return err
		}

		if hasVariable(sv) {

			sv = dotFix(sv)

			tpl, err := template.New(k).Option("missingkey=error").Funcs(funcMap).Parse(sv)
			if err != nil {
				return err
			}

			ssMap[k] = tpl
		} else {
			r.vars[k] = sv
		}
	}

	if len(ssMap) == 0 {
		return nil
	}
	buf := bytes.NewBuffer(nil)

	i := 0
	maxLoop := 10

	for ; i < maxLoop && len(ssMap) > 0; i++ {
		mapLen := len(ssMap)

		for k, tpl := range ssMap {
			buf.Reset()

			if err = tpl.Execute(buf, r.vars); err != nil {
				log.Warnf("fail to execute tpl, buf: %s, err: %s",
					buf.String(), err.Error())
				continue
			}

			val := buf.String()

			if strings.IndexRune(val, '\n') != -1 {
				log.Infof("parsed in round %v %v:\n%v", i, k, val)
			} else {
				log.Infof("parsed in round %v %v: %v", i, k, val)
			}

			r.vars[k] = val

			delete(ssMap, k)
		}

		if mapLen == len(ssMap) {
			return fmt.Errorf("can not resolve variables: %#v", ssMap)
		}
	}

	if i == maxLoop {
		return fmt.Errorf("can not resolve variables: %#v", ssMap)
	}

	return nil
}

// LoadVars 从文件中获取变量值，并set到r的变量中
func LoadVars(fn string, r *CompileT) error {
	log.Infof("load vars from file %s", fn)
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	err = LoadVarsFromString(string(dat), r)
	if err != nil {
		log.Errorf("error load vars from file %v, err : %s", fn, err.Error())
		return err
	}

	return nil
}

func extractStringMap(rf reflect.Value) map[string]interface{} {
	res := make(map[string]interface{})

	for _, key := range rf.MapKeys() {
		var k string

		switch f := key; key.Kind() {
		case reflect.Interface:
			value := f.Elem()
			if value.Kind() != reflect.String {
				return nil
			}

			k = value.String()

		case reflect.String:
			k = f.String()

		default:
			return nil
		}

		res[k] = rf.MapIndex(key).Interface()

		//log.Infof("map add string key: %#v %#v", k, res[k])
	}

	return res
}

func extractString(k string, v interface{}, r *CompileT) (string, error) {
	if sv, ok := v.(string); ok {
		return sv, nil
	}

	rf := reflect.ValueOf(v)

	var nv interface{}

	if r.Env != "" && rf.Kind() == reflect.Map {
		if sm := extractStringMap(rf); sm != nil {
			//var logMsg string

			for _, mk := range []string{"." + r.Env, "default"} {
				if mv, ok := sm[mk]; ok {
					nv = mv
					//		logMsg = fmt.Sprintf("using %v[%v] for env %v: %v", k, mk, r.Env, nv)
					break
				}
			}

			if nv != nil {
				for mk := range sm {
					if mk == "default" || mk[0] == '.' {
						continue
					}

					nv = nil

					log.Infof("will not use env for %v, because of key %v", k, mk)

					break
				}

				//if nv != nil {
				//	log.Info(logMsg)
				//}
			}
		}
	}

	if nv == nil {
		nv = v
	}

	if sv, ok := nv.(string); ok {
		return sv, nil
	}

	dat, err := yaml.Marshal(nv)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(dat)), nil
}

func dotFix(text string) string {
	res := boolRegexp.ReplaceAllString(text, "{{$1}}")
	res = noDotRegexp.ReplaceAllString(res, "{{.$1}}")

	res = indentRegexp.ReplaceAllStringFunc(res, func(s string) string {
		fs := indentRegexp.FindStringSubmatch(s)
		log.Errorf("capture indent: %v %#v", s, fs)
		return fmt.Sprintf("{{indent .%v %v}}", fs[2], len(fs[1]))
	})

	res = encryptRegexp.ReplaceAllString(res, "{{ encrypt \"$1\" }}")

	return res
}

// ExecTemplateString 从字符串执行模板替换
func ExecTemplateString(str string, r *CompileT) (string, error) {
	txt := dotFix(str)

	log.Infof("try parse template")

	tpl, err := template.New("tmpl").Option("missingkey=error").Funcs(funcMap).Parse(txt)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)

	if err := tpl.Execute(buf, r.vars); err != nil {
		return "", err
	}

	txt = buf.String()

	if strings.Index(txt, "<no value>") != -1 {

	}
	log.Infof("parse template finished")
	return buf.String(), nil
}

func indentFilterFn(in *pongo2.Value, params *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	val := in.String()
	log.Infof(val)
	return pongo2.AsValue(indent(val, 2)), nil
}

func encryptFilterFn(in *pongo2.Value, params *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	val := in.String()
	out := encrypt(val)
	return pongo2.AsValue(out), nil
}

func init() {
	pongo2.RegisterFilter("indent", indentFilterFn)
	pongo2.RegisterFilter("encrypt", encryptFilterFn)
}

func ExecPongo2TemplateString(str string, r *CompileT) (string, error) {

	tplCtx := pongo2.Context{}
	for k, v := range r.vars {
		tplCtx[k] = v
	}
	tpl, err := pongo2.FromString(str)
	if err != nil {
		return "", err
	}
	txt, err := tpl.Execute(tplCtx)
	return txt, err
}

// ExecTemplateFile 从文件执行模板替换，输入是r中的变量
func ExecTemplateFile(fn string, r *CompileT) (string, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	return ExecTemplateString(string(dat), r)
}
