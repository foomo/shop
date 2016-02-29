package shop

import (
	"encoding/xml"

	"git.bestbytes.net/Project-Globus-Services/types"
)

type ATPRequest struct {
	XMLName xml.Name       `xml:"sap-skip:http://globus.ch/xi/webshop/atp MT_ATP_REQUEST"`
	Items   []*RequestItem `xml:"ITEM,omitempty"`
}

type ATPResponse struct {
	XMLName xml.Name        `xml:"http://globus.ch/xi/webshop/atp MT_ATP_RESPONSE"`
	Items   []*ResponseItem `xml:"ITEM,omitempty"`
}

type RequestItem struct {
	XMLName xml.Name `xml:"ITEM"`

	// Positionsnummer (fortlaufende Nummerierung innerhalb des Auftrags)
	PositionNumber int32 `xml:"positionsnummer,attr,omitempty" validate:"min=1,max=999999"` // required

	// Artikelnummer (SAP)
	ItemNumber string `xml:"artikelnummer,attr,omitempty" validate:"nonzero, max=18"` // required

	// Wunschmenge
	DesiredQuantity float64 `xml:"wunschmenge,attr,omitempty" validate:"totalDigits=16,fractionDigits=3"` // required

	// Mengeneiheit (ISO bzw. SAP)
	QuantityUnit string `xml:"mengeneinheit,attr,omitempty" validate:”max=3"` // not required

	// Wunschlieferdatum
	DesiredDeliveryDate *types.XsdDateDay `xml:"wunschlieferdatum,attr,omitempty"` // not required

	// Lieferbetrieb (VZ oder Filiale)
	ShippingProvider string `xml:"lieferbetrieb,attr,omitempty" validate:”max=4"` // not required

	// PIM-Set-Nummer
	//PIMNumber string `xml:"pim_setnummer,attr,omitempty" validate:”max=18"` // not required

	// PX-Konfigurationsnummer (Möbelkonfigurator)
	//PxConfigNumber string `xml:"px_konfigurationsnummer,attr,omitempty" validate:”max=10"` // not required

	CustomData interface{}
}

type ResponseItem struct {
	XMLName xml.Name `xml:"ITEM"`

	// Positionsnummer (fortlaufende Nummerierung innerhalb des Auftrags)
	PositionNumber int32 `xml:"positionsnummer,attr,omitempty" validate:"min=1,max=999999"` // required

	// Artikelnummer (SAP)
	ItemNumber string `xml:"artikelnummer,attr,omitempty" validate:"nonzero, max=18"` // required

	// Wunschmenge
	//DesiredQuantity float64 `xml:"wunschmenge,attr,omitempty" validate:"totalDigits=16,fractionDigits=3"` // not required

	// Bestätigte Menge (ATP)
	ApprovedQuantity float64 `xml:"bestaetigte_menge,attr,omitempty" validate:"totalDigits=16,fractionDigits=3"` // required

	// Mengeneiheit (ISO bzw. SAP) (z.B. PCE, CU, KG, CS, etc.)
	QuantityUnit string `xml:"mengeneinheit,attr,omitempty" validate:"max=3"` // not required

	// Wunschlieferdatum
	DesiredDeliveryDate *types.XsdDateDay `xml:"wunschlieferdatum,attr,omitempty"` // not required

	// Lieferdatum 1: Datum an dem die Wunschmenge dem Kunden übergeben werden kann
	DeliveryDate *types.XsdDateDay `xml:"lieferdatum_1,attr,omitempty" validate:"nonzero"` // required

	// Lieferdatum 2: Nächst mögliches Lieferdatum (enthält Terminaufschlag für Filialabholung bzw. Heimlieferung
	//DeliveryDate_2 *types.XsdDateDay `xml:"lieferdatum_2,attr,omitempty"` // not required

	// Lieferbetrieb
	ShippingProvider string `xml:"lieferbetrieb,attr,omitempty" validate:”min=1, max=4"` // required

	// Bestandsbetrieb
	//StockProvider string `xml:"bestandsbetrieb,attr,omitempty" validate:"max=4"` // not required

	// PIM-Set-Nummer
	//PIMNumber string `xml:"pim_setnummer,attr,omitempty" validate:"max=18"` // not required

	// PX-Konfigurationsnummer (Möbelkonfigurator)
	//PxConfigNumber string `xml:"px_konfigurationsnummer,attr,omitempty" validate:"max=10"` // not required

	// Sonderbeschaffungsart (00=keine Sonderbeschaffung. 01=Clearing Filiale->VZ, 02=Clearing VZ->Filiale)
	//SpecialAquisitionMode string `xml:"sonderbeschaffungsart,attr,omitempty" validate:"max=2"` // not required

	// Fehlercodes 00 ... 99
	ErrorCode string `xml:"fehlercode,attr,omitempty" validate:"max=2"` // not required

	CustomData interface{}
}
