package openapi_test

import (
	"bytes"
	middlewareOapi "kafka-polygon/pkg/middleware/fiber/openapi"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var testXMLSchema = `openapi: "3.0.0"
info:
  version: 1.0.0
  title: TestServer
servers:
  - url: http://deepmap.ai
paths:
  /resource_xml:
    post:
      operationId: postResourceXml
      requestBody:
        required: true
        content:
          text/xml:
            schema:
              $ref: '#/components/schemas/submit_transfer_status_envelop_input'
      responses:
        '200':
          description: OK
        '409':
          description: Validation error
components:
  schemas:
    submit_transfer_status_envelop_input:
      type: object
      nullable: true
      xml:
        name: 'Envelope'
        prefix: 'soap'
        namespace: 'http://schemas.xmlsoap.org/soap/envelope/'
        wrapped: true 
      examples: |
        <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
          <soap:Body>
              <SubmitStatusReq xmlns="http://payment.blue.pl/bluecash/firm/types">
                  <TransferStatus>
                      <TransferAppId>a22451ab-74cb-4dd8-a7cb-3a1246a3c7b0</TransferAppId>
                      <TransferBCId>KL2Y6DAE</TransferBCId>
                      <Amount>100.00</Amount>
                      <Currency>PLN</Currency>
                      <Status>1</Status>
                      <StatusDate>2020-10-11T16:28:22.703+02:00</StatusDate>
                  </TransferStatus>
                  <Hash>3f562caf96ebbd599106ef63f7529ec58e2a5aa372ccf5988884855504560b81</Hash>
              </SubmitStatusReq>
          </soap:Body>
        </soap:Envelope>
`

func TestCustomBodyDecoder(t *testing.T) {
	t.Parallel()

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(testXMLSchema))
	require.NoError(t, err, "Error initializing swagger")

	app := fiber.New()
	swagger.Servers = nil
	app.Use(middlewareOapi.OapiRequestValidator(swagger))

	dec := middlewareOapi.NewCustomBodyDecoders()
	middlewareOapi.RegisterAdditionalBodyDecoders(dec)

	called := false

	app.Post("/resource_xml", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	t.Run("Send a good request body xml", func(t *testing.T) {
		bodyXML := []byte(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
          <soap:Body>
              <SubmitStatusReq xmlns="http://payment.blue.pl/bluecash/firm/types">
                  <TransferStatus>
                      <TransferAppId>a22451ab-74cb-4dd8-a7cb-3a1246a3c7b0</TransferAppId>
                      <TransferBCId>KL2Y6DAE</TransferBCId>
                      <Amount>100.00</Amount>
                      <Currency>PLN</Currency>
                      <Status>1</Status>
                      <StatusDate>2020-10-11T16:28:22.703+02:00</StatusDate>
                  </TransferStatus>
                  <Hash>3f562caf96ebbd599106ef63f7529ec58e2a5aa372ccf5988884855504560b81</Hash>
              </SubmitStatusReq>
          </soap:Body>
        </soap:Envelope>`)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_xml", bytes.NewBuffer(bodyXML))
		req.Header.Add("Content-Type", "text/xml")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	})

	t.Run("Send a bad request body xml", func(t *testing.T) {
		bodyXML := []byte(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
          <soap:Body>
              <SubmitStatusReq xmlns="http://payment.blue.pl/bluecash/firm/types">
                  <TransferStatus>
                      <TransferAppId>a22451ab-74cb-4dd8-a7cb-3a1246a3c7b0</TransferAppId>
                      <TransferBCId>KL2Y6DAE</TransferBCId>
                      <Amount>100.00</Amount>
                      <Currency>PLN</Currency>
                      <Status>1</Status>
                      <StatusDate>2020-10-11T16:28:22.703+02:00</StatusDate>
                  </TransferStatus>
                  <Hash>3f562caf96ebbd599106ef63f7529ec58e2a5aa372ccf5988884855504560b81</Hash>
              </SubmitStatusReq>
          </soap:Body
        </soap:Envelope>`)
		req := httptest.NewRequest(fiber.MethodPost, "/resource_xml", bytes.NewBuffer(bodyXML))
		req.Header.Add("Content-Type", "text/xml")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.False(t, called, "Handler should have been called")
		called = false
	})
}
