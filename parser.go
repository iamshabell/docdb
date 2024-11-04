package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func lexString(input []rune, index int) (string, int, error) {
	if index >= len(input) {
		return "", index, nil
	}

	if input[index] == '"' {
		index++
		eof := false

		var str []rune

		for index < len(input) {
			if input[index] == '"' {
				eof = true
				break
			}

			str = append(str, input[index])
		}

		if !eof {
			return "", index, fmt.Errorf("Expected end quoute at index %d", index)
		}

	}

	var str []rune
	var char rune

	for index < len(input) {
		char = input[index]
		if !(unicode.IsLetter(char) || unicode.IsDigit(char) || char == '.') {
			break
		}
		str = append(str, char)
		index++
	}

	if len(str) == 0 {
		return "", index, fmt.Errorf("Expected string at index %d", index)
	}

	return string(str), index, nil

}

type queryComparison struct {
	key      []string
	literal  string
	operator string
}

type query struct {
	comparisons []queryComparison
}

func parseQuery(q string) (*query, error) {
	println(q)
	if q == "" {
		return &query{}, nil
	}

	i := 0
	var parsed query
	var queryRunes = []rune(q)

	for i < len(queryRunes) {
		for unicode.IsSpace(queryRunes[i]) {
			i++
		}

		key, next, err := lexString(queryRunes, i)
		if err != nil {
			return nil, fmt.Errorf("Error parsing got [%s]: %s", err, q[next:])
		}

		if q[next] != ':' {
			return nil, fmt.Errorf("Expected colon at index %d, got: [$s]", next, q[next:])
		}
		i = next + 1

		operator := "="
		if q[i] == '>' || q[i] == '<' {
			operator = string(q[i])
			i++
		}

		literal, next, err := lexString(queryRunes, i)
		if err != nil {
			return nil, fmt.Errorf("Error parsing got [%s]: %s", err, q[next:])
		}
		i = next

		arg := queryComparison{
			key:      strings.Split(key, "."),
			literal:  literal,
			operator: operator,
		}

		parsed.comparisons = append(parsed.comparisons, arg)
	}

	return &parsed, nil
}

func getPath(doc map[string]any, parts []string) (any, bool) {
	var docSegment any = doc
	for _, part := range parts {
		m, ok := docSegment.(map[string]any)
		if !ok {
			return nil, false
		}

		if docSegment, ok = m[part]; !ok {
			return nil, false
		}
	}

	return docSegment, true
}

func (q query) match(doc map[string]any) bool {
	for _, argument := range q.comparisons {
		value, ok := getPath(doc, argument.key)
		if !ok {
			return false
		}

		if argument.operator == "=" {
			match := fmt.Sprintf("%v", value) == argument.literal
			if !match {
				return false
			}

			continue
		}

		right, err := strconv.ParseFloat(argument.literal, 64)
		if err != nil {
			return false
		}

		var left float64
		switch t := value.(type) {
		case float64:
			left = t
		case float32:
			left = float64(t)
		case uint:
			left = float64(t)
		case uint8:
			left = float64(t)
		case uint16:
			left = float64(t)
		case uint32:
			left = float64(t)
		case uint64:
			left = float64(t)
		case int:
			left = float64(t)
		case int8:
			left = float64(t)
		case int16:
			left = float64(t)
		case int32:
			left = float64(t)
		case int64:
			left = float64(t)
		case string:
			left, err = strconv.ParseFloat(t, 64)
			if err != nil {
				return false
			}
		default:
			return false
		}

		if argument.operator == ">" {
			if left <= right {
				return false
			}

			continue
		}

		if left >= right {
			return false
		}
	}

	return true
}
