package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for our application.
// We use struct tags to define environment variable names,
// whether they're required, and default values.
type Config struct {
	Port      string `env:"PORT,default=8080"`
	WebUIPath string `env:"WEBUI_PATH,default=./static"`

	TitleDumpURL string        `env:"TITLE_DUMP_URL,required"`
	FetchTimeout time.Duration `env:"FETCH_TIMEOUT,default=30s"`

	MeilisearchURL string `env:"MEILISEARCH_URL,required"`
	MeilisearchKey string `env:"MEILISEARCH_KEY,required"`
	IndexName      string `env:"INDEX_NAME,default=titles"`

	TaskTimeout time.Duration `env:"TASK_TIMEOUT,default=0"`
}

// Load populates the given struct pointer with values from environment variables.
// It returns a single error that collates all errors found during loading.
func Load(cfg interface{}) error {
	var missingVars []string
	var allErrors []string

	val := reflect.ValueOf(cfg).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip field if no tag or invalid
		tag := field.Tag.Get("env")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse the tag into parts: name, options
		parts := strings.Split(tag, ",")
		envVarName := parts[0]

		// Parse options from the tag: only "required" and "default" are valid
		isRequired := false
		defaultValue := ""
		for _, part := range parts[1:] {
			if part == "required" {
				isRequired = true
			} else if strings.HasPrefix(part, "default=") {
				defaultValue = strings.TrimPrefix(part, "default=")
			}
		}

		envValue := os.Getenv(envVarName)

		finalValue := ""
		if envValue == "" {
			if isRequired {
				missingVars = append(missingVars, envVarName)
				continue
			}
			finalValue = defaultValue
		} else {
			finalValue = envValue
		}

		// Skip setting the value if it's empty (e.g., an optional field with no default)
		if finalValue == "" {
			continue
		}

		err := setField(fieldValue, finalValue)
		if err != nil {
			errMessage := fmt.Sprintf(
				"error parsing %s for field %s: %v", envVarName, field.Name,
				err,
			)
			allErrors = append(allErrors, errMessage)
		}
	}

	// Aggregate and return errors
	if len(missingVars) > 0 {
		errMessage := fmt.Sprintf(
			"required environment variables not set or empty: %s",
			strings.Join(missingVars, ", "),
		)
		allErrors = append(allErrors, errMessage)
	}

	if len(allErrors) > 0 {
		return fmt.Errorf(
			"configuration errors:\n- %s", strings.Join(allErrors, "\n- "),
		)
	}

	return nil
}

// setField converts the string value to the field's type and sets it.
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration specifically, as it's an int64 underneath
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			// Handle regular integers
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
