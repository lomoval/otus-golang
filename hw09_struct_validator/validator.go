package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	tagNameValidate = "validate"
	tagValueNested  = "nested"
	tagValueIn      = "in"
	tagValueMax     = "max"
	tagValueMin     = "min"
	tagValueLen     = "len"
	tagValueRegexp  = "regexp"
)

var (
	ErrIncorrectTagValue        = errors.New("incorrect tag value for validating with field value")
	ErrValidateIncorrectLen     = errors.New("value has incorrect length")
	ErrValidateNotMatchRegexp   = errors.New("does not match regexp")
	ErrValidateNotFoundInList   = errors.New("does not found in list")
	ErrValidateIncorrectNumeric = errors.New("incorrect numeric value")
	ErrIncorrectTag             = errors.New("incorrect tag")
	ErrIncorrectStruct          = errors.New("incorrect struct")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

type validationData struct {
	Name            string
	SourceValue     string
	ConvertedValues interface{}
}

func (v ValidationErrors) Error() string {
	sort.Slice(v, func(i, j int) bool {
		if v[i].Field == v[j].Field {
			return v[i].Err.Error() < v[j].Err.Error()
		}
		return v[i].Field < v[j].Field
	})
	b := strings.Builder{}
	for _, validationError := range v {
		b.WriteString(fmt.Sprintf("{name: %s, error: %s}", validationError.Field, validationError.Err.Error()))
	}
	return b.String()
}

func Validate(v interface{}) error {
	if v == nil {
		return ErrIncorrectStruct
	}
	rv := reflect.ValueOf(v)
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct {
		return ErrIncorrectStruct
	}

	var validatorErrors ValidationErrors
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		validationData, err := parseValidateTag(field, t.Field(i).Tag)
		if err != nil {
			return err
		}
		if len(validationData) == 0 {
			continue
		}

		if validationData[0].Name == tagValueNested {
			err = Validate(field.Interface())
			var vErrors ValidationErrors
			if errors.As(err, &vErrors) {
				validatorErrors = append(validatorErrors, vErrors...)
				continue
			}
			return err
		}

		validatorErrors, err = validateField(validationData, field, t.Field(i).Name, validatorErrors)
		if err != nil {
			return err
		}
	}

	if len(validatorErrors) == 0 {
		return nil
	}
	return validatorErrors
}

func validateField(
	validationData []validationData,
	field reflect.Value,
	fieldName string,
	validatorErrors ValidationErrors,
) (ValidationErrors, error) {
	var err error

	if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
		for _, data := range validationData {
			validatorErrors, err = validateSlice(field, fieldName, data, validatorErrors)
			if err != nil {
				return nil, err
			}
		}
		return validatorErrors, nil
	}

	for _, data := range validationData {
		if validatorErrors, err = validateValue(
			fieldName,
			field.Interface(),
			field.Type(),
			data,
			validatorErrors,
		); err != nil {
			return nil, err
		}
	}
	return validatorErrors, nil
}

func validateSlice(
	field reflect.Value,
	fieldName string,
	data validationData,
	validatorErrors ValidationErrors,
) (ValidationErrors, error) {
	var err error
	for f := 0; f < field.Len(); f++ {
		validatorErrors, err = validateValue(
			fieldName,
			field.Index(f).Interface(),
			field.Type().Elem(),
			data,
			validatorErrors,
		)
		if err != nil {
			return nil, err
		}
	}
	return validatorErrors, nil
}

func validateValue(
	fieldName string,
	val interface{},
	t reflect.Type,
	data validationData,
	validationErrors ValidationErrors,
) (ValidationErrors, error) {
	if t.Kind() == reflect.String {
		var err error
		validationErrors, err = validateStringValue(fieldName, val, data, validationErrors)
		if err != nil {
			return nil, err
		}
	}

	if data.Name == tagValueIn {
		if !inSlice(val, data.ConvertedValues) {
			validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: ErrValidateNotFoundInList})
		}
		return validationErrors, nil
	}

	if data.Name == tagValueMin {
		if !ge(val, reflect.ValueOf(data.ConvertedValues).Index(0).Interface()) {
			validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: ErrValidateIncorrectNumeric})
		}
		return validationErrors, nil
	}

	if data.Name == tagValueMax {
		if !le(val, reflect.ValueOf(data.ConvertedValues).Index(0).Interface()) {
			validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: ErrValidateIncorrectNumeric})
		}
		return validationErrors, nil
	}
	return validationErrors, nil
}

func validateStringValue(
	fieldName string,
	val interface{},
	data validationData,
	validationErrors ValidationErrors,
) (ValidationErrors, error) {
	if data.Name == tagValueLen {
		check, err := strconv.Atoi(data.SourceValue)
		if err != nil {
			return nil, ErrIncorrectTagValue
		}
		if len(val.(string)) != check {
			validationErrors = append(validationErrors, ValidationError{
				Field: fieldName,
				Err:   ErrValidateIncorrectLen,
			})
		}
		return validationErrors, nil
	}

	if data.Name == tagValueRegexp {
		r, err := regexp.Compile(data.SourceValue)
		if err != nil {
			return nil, ErrIncorrectTagValue
		}
		match := r.FindString(val.(string))
		if len(match) != len(val.(string)) {
			return append(validationErrors, ValidationError{Field: fieldName, Err: ErrValidateNotMatchRegexp}), nil
		}
	}

	return validationErrors, nil
}

func parseValidateTag(field reflect.Value, tag reflect.StructTag) ([]validationData, error) {
	val := tag.Get(tagNameValidate)
	if val == "" {
		return nil, nil
	}
	validators := strings.Split(val, "|")

	data := make([]validationData, 0, len(validators))
	for _, validator := range validators {
		validatorParts := strings.SplitN(validator, ":", 2)
		if len(validatorParts) == 1 {
			if validatorParts[0] != tagValueNested {
				return nil, ErrIncorrectTag
			}
			// Ignore other validators if nested
			return []validationData{{Name: validatorParts[0]}}, nil
		}

		validatorValues := []string{validatorParts[1]}
		if validatorParts[0] == tagValueIn {
			validatorValues = strings.Split(validatorParts[1], ",")
		}

		fieldType := field.Type()
		if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			fieldType = field.Type().Elem()
		}

		values, err := convertSliceByType(validatorValues, fieldType)
		if err != nil {
			return nil, err
		}
		data = append(data, validationData{Name: validatorParts[0], ConvertedValues: values, SourceValue: validatorParts[1]})
	}
	return data, nil
}

func convertSliceByType(values []string, t reflect.Type) (interface{}, error) {
	slice := reflect.MakeSlice(reflect.SliceOf(t), 0, len(values))
	size := (int)(t.Size()) * 8
	var convert func(v string) (interface{}, error)

	// Not planning to implement converters for all types
	//exhaustive:ignore
	switch t.Kind() {
	case reflect.String:
		if t.String() == "string" {
			return values, nil
		}
		convert = func(str string) (interface{}, error) {
			productPointer := reflect.New(t).Elem()
			productPointer.SetString(str)
			return productPointer.Interface(), nil
		}
	case reflect.Int8:
		convert = func(str string) (interface{}, error) {
			val, err := strconv.ParseInt(str, 0, size)
			return (int8)(val), err
		}
	case reflect.Int16:
		convert = func(str string) (interface{}, error) {
			val, err := strconv.ParseInt(str, 0, size)
			return (int16)(val), err
		}
	case reflect.Int32:
		convert = func(str string) (interface{}, error) {
			val, err := strconv.ParseInt(str, 0, size)
			return (int32)(val), err
		}
	case reflect.Int64:
		convert = func(str string) (interface{}, error) {
			return strconv.ParseInt(str, 0, size)
		}
	case reflect.Float32:
		convert = func(str string) (interface{}, error) {
			val, err := strconv.ParseFloat(str, size)
			return (float32)(val), err
		}
	case reflect.Float64:
		convert = func(str string) (interface{}, error) {
			return strconv.ParseFloat(str, size)
		}
	default:
		convert = func(str string) (interface{}, error) {
			val, err := strconv.ParseInt(str, 0, size)
			return (int)(val), err
		}
	}

	for _, val := range values {
		v, err := convert(val)
		if err != nil {
			return nil, ErrIncorrectTagValue
		}
		slice = reflect.Append(slice, reflect.ValueOf(v))
	}
	return slice.Interface(), nil
}

func inSlice(v interface{}, in interface{}) bool {
	s := reflect.ValueOf(in)
	for i := 0; i < s.Len(); i++ {
		if v == s.Index(i).Interface() {
			return true
		}
	}
	return false
}

func le(a interface{}, b interface{}) bool {
	switch v := a.(type) {
	case int8:
		return v <= b.(int8)
	case int16:
		return v <= b.(int16)
	case int32:
		return v <= b.(int32)
	case int64:
		return v <= b.(int64)
	case float32:
		return v <= b.(float32)
	case float64:
		return v <= b.(float64)
	default:
		return a.(int) <= b.(int)
	}
}

func ge(a interface{}, b interface{}) bool {
	switch v := a.(type) {
	case int8:
		return v >= b.(int8)
	case int16:
		return v >= b.(int16)
	case int32:
		return v >= b.(int32)
	case int64:
		return v >= b.(int64)
	case float32:
		return v >= b.(float32)
	case float64:
		return v >= b.(float64)
	default:
		return a.(int) >= b.(int)
	}
}
