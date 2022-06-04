package web

import (
	"github.com/gin-gonic/gin"
	"time"
)

type QueryParamValue struct {
	value interface{}
}

type QueryParamsParser func(string) interface{}
type QueryParamsParsingRules = map[string]QueryParamsParser
type QueryParamsMap = map[string]QueryParamValue

func (queryParamValue QueryParamValue) String() string {
	s, ok := queryParamValue.value.(string)
	if ok {
		return s
	}

	return ""
}

func (queryParamValue QueryParamValue) Int() int {
	i, ok := queryParamValue.value.(int)
	if ok {
		return i
	}

	return 0
}

func (queryParamValue QueryParamValue) Bool() bool {
	b, ok := queryParamValue.value.(bool)
	if ok {
		return b
	}

	return false
}

func (queryParamValue QueryParamValue) Time() *time.Time {
	t, ok := queryParamValue.value.(*time.Time)
	if ok {
		return t
	}

	return nil
}

func ParseQueryParams(c *gin.Context, rules QueryParamsParsingRules) QueryParamsMap {
	var result = make(QueryParamsMap)

	for param, parser := range rules {
		value := c.Query(param)
		parsedValue := parser(value)
		result[param] = QueryParamValue{parsedValue}
	}

	return result
}
