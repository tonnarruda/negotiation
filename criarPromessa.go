package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	apiKey  = "0a67fd26-32c1-11ed-a261-0242ac120002"
	baseURL = "https://negotiations-api.olaisaac.dev"
)

var response AgreementSimulationResponse

type Installment struct {
	ID                          string     `json:"id"`
	Description                 string     `json:"description"`
	Type                        string     `json:"type"`
	DueDate                     string     `json:"due_date"`
	PaidDate                    string     `json:"paid_date"`
	EnabledPaymentMethods       []string   `json:"enabled_payment_methods"`
	Status                      string     `json:"status"`
	Product                     string     `json:"product"`
	BaseAmount                  int        `json:"base_amount"`
	DueAmount                   int        `json:"due_amount"`
	OriginalAmount              int        `json:"original_amount"`
	Overdue                     bool       `json:"overdue"`
	CurrentAmount               int        `json:"current_amount"`
	CurrentDiscount             int        `json:"current_discount"`
	CurrentEarlyPaymentDiscount int        `json:"current_early_payment_discount"`
	CurrentDuePaymentDiscount   int        `json:"current_due_payment_discount"`
	CurrentInterest             int        `json:"current_interest"`
	ExemptInterest              int        `json:"exempt_interest"`
	CurrentFine                 int        `json:"current_fine"`
	ExemptFine                  int        `json:"exempt_fine"`
	LostDuePaymentDiscount      int        `json:"lost_due_payment_discount"`
	LostEarlyPaymentDiscount    int        `json:"lost_early_payment_discount"`
	EarlyPaymentDiscounts       []struct{} `json:"early_payment_discounts"`
	DuePaymentDiscounts         []struct{} `json:"due_payment_discounts"`
	PerpetualDiscounts          []struct{} `json:"perpetual_discounts"`
	ValidDiscounts              int        `json:"valid_discounts"`
	InstallmentNumber           string     `json:"installment_number"`
	Agreement                   struct {
		ID                          string      `json:"id"`
		CreatedAt                   string      `json:"created_at"`
		UpdatedAt                   string      `json:"updated_at"`
		DeletedAt                   string      `json:"deleted_at"`
		ShortID                     string      `json:"short_id"`
		Type                        string      `json:"type"`
		EarlyPaymentDiscountApplied int         `json:"early_payment_discount_applied"`
		DuePaymentDiscountApplied   int         `json:"due_payment_discount_applied"`
		InterestApplied             int         `json:"interest_applied"`
		FineApplied                 int         `json:"fine_applied"`
		OriginalInstallmentIDs      []string    `json:"original_installment_ids"`
		ResultantInstallmentIDs     []string    `json:"resultant_installment_ids"`
		ResultantInvoicesIDs        []string    `json:"resultant_invoices_ids"`
		AgreementDate               string      `json:"agreement_date"`
		AgreementAmount             int         `json:"agreement_amount"`
		ExemptionReasons            interface{} `json:"exemption_reasons"`
		ExemptFine                  int         `json:"exempt_fine"`
		ExemptInterest              int         `json:"exempt_interest"`
		SchoolDiscount              int         `json:"school_discount"`
		ChosenPaymentPlan           struct {
			PaymentMethod        string `json:"payment_method"`
			DownPaymentAmount    int    `json:"down_payment_amount"`
			NumberOfInstallments int    `json:"number_of_installments"`
		} `json:"chosen_payment_plan"`
		SimulationID   string `json:"simulation_id"`
		ResultInvoices []struct {
			ID            string `json:"id"`
			PaymentLink   string `json:"payment_link"`
			Installment   string `json:"installment"`
			InstallmentID string `json:"installment_id"`
			DueDate       string `json:"due_date"`
			Amount        int    `json:"amount"`
		} `json:"result_invoices"`
		Channel string `json:"channel"`
	} `json:"agreement"`
	Invoice struct {
		ID            string `json:"id"`
		DueDate       string `json:"due_date"`
		DigitableLine string `json:"digitable_line"`
		PixCode       string `json:"pix_code"`
		PaymentLink   string `json:"payment_link"`
		BadCredit     bool   `json:"bad_credit"`
		Status        string `json:"status"`
		EmailSent     string `json:"email_sent"`
	} `json:"invoice"`
	Contract struct {
		ID                   string `json:"id"`
		ReferenceYear        string `json:"reference_year"`
		NumberOfInstallments int    `json:"number_of_installments"`
		SchoolName           string `json:"school_name"`
		ERP                  string `json:"erp"`
	} `json:"contract"`
	Student struct {
		Name     string `json:"name"`
		SourceID string `json:"source_id"`
	} `json:"student"`
	Guardian struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		TaxID       string `json:"tax_id"`
		SourceID    string `json:"source_id"`
	} `json:"guardian"`
}

type ResponseData struct {
	Data []Installment `json:"data"`
}

type AgreementSimulationResponse struct {
	Data struct {
		ID        string  `json:"id"`
		DueAmount float64 `json:"due_amount"`
	} `json:"data"`
}

func makeRequest(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-api-key", apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

func getOverdueInstallment(taxID string) string {
	url := fmt.Sprintf("%s/api/guardians/%s/installments", baseURL, taxID)

	resp, err := makeRequest("GET", url, nil)
	if err != nil {
		log.Printf("Erro ao fazer a solicitação GET: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("A solicitação retornou um status não OK: %d\n", resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler o corpo da resposta: %v\n", err)
		return ""
	}

	var responseData ResponseData
	if err := json.Unmarshal([]byte(body), &responseData); err != nil {
		log.Printf("Erro ao decodificar a resposta JSON: %v\n", err)
		return ""
	}

	for _, item := range responseData.Data {
		if item.Overdue && item.Type == "TUITION" {
			fmt.Printf("ID do Installment Vencido (TUITION): %s\n", item.ID)
			return item.ID
		}
	}
	fmt.Println("Não existem parcelas vencidas para esse RF.")
	return ""
}

func createAgreementSimulation(taxID, installmentID string) string {
	url := fmt.Sprintf("%s/api/guardians/%s/agreement-simulations", baseURL, taxID)

	requestBody := map[string]interface{}{
		"installments_ids":  []string{installmentID},
		"exemption_reasons": []string{},
		"school_discount":   0,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Erro ao criar o corpo JSON: %v\n", err)
		return ""
	}

	resp, err := makeRequest("POST", url, jsonBody)
	if err != nil {
		log.Printf("Erro ao fazer a solicitação POST: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("A solicitação retornou um status não OK: %d\n", resp.StatusCode)
		return ""
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Erro ao decodificar a resposta JSON: %v\n", err)
		return ""
	}

	fmt.Println("Simulação concluída com sucesso!")
	fmt.Printf("Agreement Simulation ID: %s\n", response.Data.ID)

	return response.Data.ID
}

func createPromisse(agreementSimulationID string, downPaymentAmount float64) {

	url := fmt.Sprintf("%s/api/v1/negotiations", baseURL)

	currentDate := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02T15:04:05.999Z")

	requestBody := map[string]interface{}{
		"agreement_simulation_id": agreementSimulationID,
		"down_payment": map[string]interface{}{
			"amount":   downPaymentAmount,
			"due_date": currentDate,
		},
		"payment_plan": []map[string]interface{}{
			{
				"methods":                      []string{"CREDIT_CARD"},
				"index":                        1,
				"max_installments_credit_card": 1,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Erro ao criar o corpo JSON:", err)
		return
	}

	resp, err := makeRequest("POST", url, jsonBody)
	if err != nil {
		log.Printf("Erro ao fazer a solicitação POST: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("A solicitação retornou um status não OK: %d\n", resp.StatusCode)
		return
	}

	fmt.Println("Promessa criada com sucesso!")

}

func main() {
	fmt.Print("Digite o CPF do RF (Apenas números): ")
	var taxID string
	fmt.Scanln(&taxID)

	installmentID := getOverdueInstallment(taxID)
	if installmentID != "" {
		createAgreementSimulation(taxID, installmentID)
		createPromisse(response.Data.ID, response.Data.DueAmount)
	}
}
