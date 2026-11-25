package contact

import "github.com/mheers/fastbill-api-golang"

// Filter narrows contact.get.
type Filter struct {
	ContactID      fastbill.ID `json:"CONTACT_ID,omitempty"`
	CustomerID     fastbill.ID `json:"CUSTOMER_ID,omitempty"`
	CustomerNumber string      `json:"CUSTOMER_NUMBER,omitempty"`
	// Term searches ORGANIZATION, FIRST_NAME, LAST_NAME, ADDRESS,
	// ADDRESS_2, ZIPCODE, EMAIL and TAGS.
	Term string `json:"TERM,omitempty"`
}

// Results of contact.update (ADDITIONAL_INFO).
const (
	// UpdateUpdatedOld means the contact was updated in place.
	UpdateUpdatedOld = "updated_old"
	// UpdateCreateNew means the contact was locked and replaced by a new
	// one; the response carries its CONTACT_ID.
	UpdateCreateNew = "create_new"
)

// Contact is a contact person as contact.get returns it, and the data of
// contact.create and contact.update.
type Contact struct {
	// ContactID is required for Update and ignored by Create.
	ContactID fastbill.ID `json:"CONTACT_ID,omitempty"`
	// CustomerID is the customer the contact belongs to; it is required.
	CustomerID     fastbill.ID `json:"CUSTOMER_ID,omitempty"`
	Organization   string      `json:"ORGANIZATION,omitempty"`
	Position       string      `json:"POSITION,omitempty"`
	AcademicDegree string      `json:"ACADEMIC_DEGREE,omitempty"`
	// Salutation is mr, mrs, family or empty.
	Salutation  string `json:"SALUTATION,omitempty"`
	FirstName   string `json:"FIRST_NAME,omitempty"`
	LastName    string `json:"LAST_NAME,omitempty"`
	Address     string `json:"ADDRESS,omitempty"`
	Address2    string `json:"ADDRESS_2,omitempty"`
	ZipCode     string `json:"ZIPCODE,omitempty"`
	City        string `json:"CITY,omitempty"`
	CountryCode string `json:"COUNTRY_CODE,omitempty"`
	// SecondaryAddress is the delivery address.
	SecondaryAddress string `json:"SECONDARY_ADDRESS,omitempty"`
	Phone            string `json:"PHONE,omitempty"`
	Phone2           string `json:"PHONE_2,omitempty"`
	Fax              string `json:"FAX,omitempty"`
	Mobile           string `json:"MOBILE,omitempty"`
	Email            string `json:"EMAIL,omitempty"`
	Website          string `json:"WEBSITE,omitempty"`
	VatID            string `json:"VAT_ID,omitempty"`
	CurrencyCode     string `json:"CURRENCY_CODE,omitempty"`
	Comment          string `json:"COMMENT,omitempty"`
	Created          string `json:"CREATED,omitempty"`
	LastUpdate       string `json:"LASTUPDATE,omitempty"`
	Tags             string `json:"TAGS,omitempty"`
}

// CreateResponse is the answer of contact.create.
type CreateResponse struct {
	Status    string      `json:"STATUS"`
	ContactID fastbill.ID `json:"CONTACT_ID"`
}

// UpdateResponse is the answer of contact.update. AdditionalInfo is
// UpdateUpdatedOld or UpdateCreateNew; ContactID is only set when the
// contact was replaced.
type UpdateResponse struct {
	Status         string      `json:"STATUS"`
	AdditionalInfo string      `json:"ADDITIONAL_INFO"`
	ContactID      fastbill.ID `json:"CONTACT_ID"`
}

type idRequest struct {
	CustomerID fastbill.ID `json:"CUSTOMER_ID"`
	ContactID  fastbill.ID `json:"CONTACT_ID"`
}

type getResponse struct {
	Contacts fastbill.List[Contact] `json:"CONTACTS"`
}
