package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Form struct{
	url.Values
	Errors errors
}

// return true if there are no errors , otherwise false
func (form *Form) Valid() bool {
	return len(form.Errors) == 0
}

//New : initialize a form struct
func New (data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}
// Has checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	if x == "" {
		return false
	}
	return true
}

//required filed to fill
func (f *Form) Required(fields ...string){
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == ""{
			f.Errors.Add(field, "The field can not be empty")

		}
	}
}

//check for string minimum lenght
func (f *Form) MinLenght (field string , length int , r *http.Request) bool{
	x:= r.Form.Get(field)
	if len(x) < length {
		f.Errors.Add(field,fmt.Sprintf("This field must be at least %d charachter long",length))
		return false
	}
	return true
}

//check email validation with go validator package
func (f * Form) IsEmail(field string){
	if !govalidator.IsEmail(f.Get(field)){
		f.Errors.Add(field,"Invalid Email address")
	}
}