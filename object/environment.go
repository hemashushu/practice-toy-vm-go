// original from https://interpreterbook.com/

package object

type Environment struct {
	store map[string]Object // records
	outer *Environment      // 上一层环境
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]

	// 向上一层获取标识符的值
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}

	return obj, ok
}

// 定义标识符
func (e *Environment) Set(name string, value Object) Object {
	e.store[name] = value
	return value
}
