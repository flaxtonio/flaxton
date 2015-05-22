package consoleWizard
import "encoding/json"

type WizardEndsFunction func(w *Wizard)
type WizardEndsJsonFunction func(json_str string)
type WizardErrorFunction func(e error)

type WizardConsole struct {
	WizardFields Wizard
	JsonText string
	FinishCallback WizardEndsFunction
	FinishJsonCallback WizardEndsJsonFunction
	ErrorCallback WizardErrorFunction
}

func InitWizardConsole(json_fields string) (wc WizardConsole) {
	wc.JsonText = json_fields
	wc.ErrorCallback = func(err error) {}
	wc.FinishCallback = func(w *Wizard) {}
	wc.FinishJsonCallback = func(json_str string) {}
	return wc
}

func (wc *WizardConsole) RunConsole() {
	p_error := wc.WizardFields.HandleJsonFields(wc.JsonText)
	if p_error != nil {
		wc.ErrorCallback(p_error)
	}
	for _, f := range wc.WizardFields.Fields  {
		switch f.Type {
			case Number:
				{
					field := f.Content.(FieldNumber)
					field.HandleConsole()
				}
			case Text:
				{
					field := f.Content.(FieldText)
					field.HandleConsole()
				}
			case MultiInput:
				{
					field := f.Content.(FieldMultiInput)
					field.HandleConsole()
				}
			case Choice:
				{
					field := f.Content.(FieldChoice)
					field.HandleConsole()
				}
		}
	}

	go wc.FinishCallback(&wc.WizardFields)
	json_str, err := json.Marshal(wc.WizardFields)
	if err != nil {
		wc.ErrorCallback(err)
	}
	go wc.FinishJsonCallback(string(json_str))
}