// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/kaz-as/test-transactions/models"
)

// CreateUserOKCode is the HTTP code returned for type CreateUserOK
const CreateUserOKCode int = 200

/*
CreateUserOK user created

swagger:response createUserOK
*/
type CreateUserOK struct {

	/*
	  In: Body
	*/
	Payload *models.CreateUserSuccess `json:"body,omitempty"`
}

// NewCreateUserOK creates CreateUserOK with default headers values
func NewCreateUserOK() *CreateUserOK {

	return &CreateUserOK{}
}

// WithPayload adds the payload to the create user o k response
func (o *CreateUserOK) WithPayload(payload *models.CreateUserSuccess) *CreateUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create user o k response
func (o *CreateUserOK) SetPayload(payload *models.CreateUserSuccess) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*
CreateUserDefault generic error response

swagger:response createUserDefault
*/
type CreateUserDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateUserDefault creates CreateUserDefault with default headers values
func NewCreateUserDefault(code int) *CreateUserDefault {
	if code <= 0 {
		code = 500
	}

	return &CreateUserDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the create user default response
func (o *CreateUserDefault) WithStatusCode(code int) *CreateUserDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the create user default response
func (o *CreateUserDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the create user default response
func (o *CreateUserDefault) WithPayload(payload *models.Error) *CreateUserDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create user default response
func (o *CreateUserDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateUserDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
