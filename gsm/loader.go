package gsm

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Loader loads configuration into a struct using field tags.
type Loader struct {
	resolver *Resolver
}

// LoaderOption is a functional option for configuring a Loader.
type LoaderOption = ResolverOption

// NewLoader creates a new Loader with the given client and options.
// The client can be nil if Secret Manager is not used.
//
// Example:
//
//	loader := gsm.NewLoader(client,
//	    gsm.WithEnvPrefix("APP_"),
//	    gsm.WithSecretManagerEnabled(true),
//	)
func NewLoader(client *Client, opts ...LoaderOption) *Loader {
	return &Loader{
		resolver: NewResolver(client, opts...),
	}
}

// Load loads configuration values into the provided struct pointer.
// The struct fields should be tagged with `gsm:"SECRET_NAME,option1,option2"`.
//
// Supported tag options:
//   - "SECRET_NAME" - The name of the environment variable/secret (required)
//   - "default=VALUE" - Default value if not found
//   - "required" - Returns error if value is not found
//   - "-" - Skip this field
//
// Supported field types:
//   - string
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//   - []string
//
// Example:
//
//	type Config struct {
//	    APIKey string   `gsm:"API_KEY,required"`
//	    DBHost string   `gsm:"DB_HOST,default=localhost"`
//	    DBPort int      `gsm:"DB_PORT,default=5432"`
//	    Debug  bool     `gsm:"DEBUG,default=false"`
//	}
//
//	var cfg Config
//	err := loader.Load(ctx, &cfg)
func (l *Loader) Load(ctx context.Context, target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return ErrInvalidTarget
	}

	return l.loadStruct(ctx, v.Elem())
}

func (l *Loader) loadStruct(ctx context.Context, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("gsm")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse tag
		tagInfo := parseTag(tag)
		if tagInfo.secretName == "" {
			continue
		}

		// Build the reference string
		var refString string
		if tagInfo.hasDefault {
			refString = fmt.Sprintf("sm://%s||%s", tagInfo.secretName, tagInfo.defaultValue)
		} else {
			refString = fmt.Sprintf("sm://%s", tagInfo.secretName)
		}

		// Resolve and set the value
		if err := l.setField(ctx, field, fieldType, refString); err != nil {
			if tagInfo.required {
				return &RequiredFieldError{
					FieldName:  fieldType.Name,
					SecretName: tagInfo.secretName,
				}
			}
			// If not required and there's an error, continue with next field
			continue
		}
	}

	return nil
}

func (l *Loader) setField(ctx context.Context, field reflect.Value, fieldType reflect.StructField, refString string) error {
	switch field.Kind() {
	case reflect.String:
		value, err := l.resolver.Resolve(ctx, refString)
		if err != nil {
			return err
		}
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := l.resolver.Resolve(ctx, refString)
		if err != nil {
			return err
		}
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int for field %s: %w", fieldType.Name, err)
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := l.resolver.Resolve(ctx, refString)
		if err != nil {
			return err
		}
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse uint for field %s: %w", fieldType.Name, err)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		value, err := l.resolver.Resolve(ctx, refString)
		if err != nil {
			return err
		}
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float for field %s: %w", fieldType.Name, err)
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		value, err := l.resolver.Resolve(ctx, refString)
		if err != nil {
			return err
		}
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse bool for field %s: %w", fieldType.Name, err)
		}
		field.SetBool(boolVal)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// For []string, resolve as a slice
			values, err := l.resolver.ResolveSlice(ctx, []string{refString})
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(values))
		} else {
			return &UnsupportedTypeError{
				FieldName: fieldType.Name,
				TypeName:  field.Type().String(),
			}
		}

	default:
		return &UnsupportedTypeError{
			FieldName: fieldType.Name,
			TypeName:  field.Type().String(),
		}
	}

	return nil
}

type tagInfo struct {
	secretName   string
	defaultValue string
	hasDefault   bool
	required     bool
}

// parseTag parses a struct tag in the format: "SECRET_NAME,default=value,required"
func parseTag(tag string) tagInfo {
	parts := strings.Split(tag, ",")
	info := tagInfo{
		secretName: strings.TrimSpace(parts[0]),
	}

	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])

		if part == "required" {
			info.required = true
		} else if strings.HasPrefix(part, "default=") {
			info.defaultValue = strings.TrimPrefix(part, "default=")
			info.hasDefault = true
		}
	}

	return info
}
