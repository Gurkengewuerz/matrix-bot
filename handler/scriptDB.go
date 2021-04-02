package handler

import (
	"github.com/dop251/goja"
	"strings"
)

func (pm *PluginHandler) scriptDBQuery(sql goja.Value, args ...goja.Value) goja.Value {
	pm.Logger.WithField("script", pm.currentPlugin.name).Debug("scriptDBQuery()")
	if len(args)%2 != 0 {
		panic(pm.vm.ToValue("args has to be of factor two"))
	}

	data := make([]interface{}, len(args)/2)

	for i := 0; i < len(args)/2; i++ {
		x := i * 2
		argType := args[x]
		value := args[x+1]

		switch strings.ToLower(argType.String()) {
		case "string":
			data[i] = value.String()
			break
		case "int":
			data[i] = value.ToInteger()
			break
		case "bool":
			data[i] = value.ToBoolean()
			break
		case "float":
			data[i] = value.ToFloat()
			break
		case "undefined":
			data[i] = nil
			break
		default:
			panic("unknown data type")
		}
	}

	rows, err := pm.DB.Query(sql.String(), data...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		result    [][]string
		container []string
		pointers  []interface{}
	)

	cols, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	length := len(cols)

	for rows.Next() {
		pointers = make([]interface{}, length)
		container = make([]string, length)

		for i := range pointers {
			pointers[i] = &container[i]
		}

		err = rows.Scan(pointers...)
		if err != nil {
			panic(err.Error())
		}

		result = append(result, container)
	}

	return pm.vm.ToValue(result)
}

func (pm *PluginHandler) scriptDBPreparedStatement(sql goja.Value, args ...goja.Value) goja.Value {
	pm.Logger.WithField("script", pm.currentPlugin.name).Debug("scriptDBPreparedStatement()")
	statement, err := pm.DB.Prepare(sql.String())
	if err != nil {
		panic(pm.vm.ToValue(err))
	}
	defer statement.Close()

	if len(args)%2 != 0 {
		panic(pm.vm.ToValue("args has to be of factor two"))
	}

	data := make([]interface{}, len(args)/2)

	for i := 0; i < len(args)/2; i++ {
		x := i * 2
		argType := args[x]
		value := args[x+1]

		switch strings.ToLower(argType.String()) {
		case "string":
			data[i] = value.String()
			break
		case "int":
			data[i] = value.ToInteger()
			break
		case "bool":
			data[i] = value.ToBoolean()
			break
		case "float":
			data[i] = value.ToFloat()
			break
		case "undefined":
			data[i] = nil
			break
		default:
			panic(pm.vm.ToValue("unknown data type"))
		}
	}

	_, err = statement.Exec(data...)
	if err != nil {
		panic(pm.vm.ToValue(err))
	}

	return goja.Undefined()
}
