package consoleWizard

import (
	"encoding/json"
)

type FieldType string

// Field Types
const (
	Number FieldType 		= "number"
	Text FieldType 			= "text"
	Choice FieldType 		= "choice"
	MultiInput FieldType 	= "multiple_input" // Multiple inputs separated by symbol
)

type WizardField struct {
	ID string				`json:"id"`
	Type FieldType 			`json:"type"`
	Content interface{} 	`json:"content"` // Object based on type, one of the Filed structures
}

type Wizard struct {
	Fields []WizardField	`json:"fields"`
}
/*
JSON example
{
	fields: [
		{
			id: "some_identifier",
			type: "number",
			content: {
				label: "Type Default TCP Port Number",
				default: 80,
				value: 0 // Value field may not contain on input
			}
		},
		****************************
		****************************
	]
}
 */

func (w *Wizard) HandleJsonFields(json_string string) (err error) {
	err = nil
	err = json.Unmarshal([]byte(json_string), &w)
	return err
}