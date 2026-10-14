package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

func gronUnsorted(r io.Reader, w io.Writer, opts int) (int, error) {
	decoder := json.NewDecoder(withoutUTF8BOM(r))
	decoder.UseNumber()
	first, err := decoder.Token()
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to decode JSON: %w", err)
	}
	conv := statementToColorString
	if opts&optMonochrome > 0 {
		conv = statementToString
	}
	buffered := bufio.NewWriter(w)
	err = walkJSON(decoder, first, statement{{"json", typBare}}, func(s statement) error {
		if opts&optJSON > 0 {
			var err error
			s, err = s.jsonify()
			if err != nil {
				return err
			}
		}
		line := conv(s)
		if buffered.Available() < len(line)+1 {
			if err := buffered.Flush(); err != nil {
				return err
			}
		}
		_, err := fmt.Fprintln(buffered, line)
		return err
	})
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to form statements: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return exitFormStatements, fmt.Errorf("multiple JSON values; use --stream")
		}
		return exitFormStatements, fmt.Errorf("invalid trailing JSON: %w", err)
	}
	if err := buffered.Flush(); err != nil {
		return exitJSONEncode, fmt.Errorf("failed to write output: %w", err)
	}
	return exitOK, nil
}

// walkJSON emits one statement at a time in input order.
func walkJSON(decoder *json.Decoder, value json.Token, path statement, emit func(statement) error) error {
	emitValue := func(value interface{}) error {
		line := make(statement, len(path), len(path)+3)
		copy(line, path)
		line = append(line, token{"=", typEquals}, valueTokenFromInterface(value), token{";", typSemi})
		return emit(line)
	}

	if delimiter, ok := value.(json.Delim); ok {
		switch delimiter {
		case '{':
			if err := emitValue(map[string]interface{}{}); err != nil {
				return err
			}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				child := path.withQuotedKey(key)
				if validIdentifier(key) {
					child = path.withBare(key)
				}
				valueToken, err := decoder.Token()
				if err != nil {
					return err
				}
				if err := walkJSON(decoder, valueToken, child, emit); err != nil {
					return err
				}
			}
			end, err := decoder.Token()
			if err != nil {
				return err
			}
			if end != json.Delim('}') {
				return fmt.Errorf("expected object end")
			}
			return nil
		case '[':
			if err := emitValue([]interface{}{}); err != nil {
				return err
			}
			for index := 0; decoder.More(); index++ {
				child := path.withNumericKey(index)
				valueToken, err := decoder.Token()
				if err != nil {
					return err
				}
				if err := walkJSON(decoder, valueToken, child, emit); err != nil {
					return err
				}
			}
			end, err := decoder.Token()
			if err != nil {
				return err
			}
			if end != json.Delim(']') {
				return fmt.Errorf("expected array end")
			}
			return nil
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
		}
	}
	return emitValue(value)
}
