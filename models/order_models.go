package models

type Address struct {
	Street string `json:"street" bson:"street"`
	City   string `json:"city" bson:"city"`
	State  string `json:"state" bson:"state"`
	Pin    string `json:"pin" bson:"pin"`
}
type CustomerDetails struct {
	Name    string  `json:"name" bson:"name"`
	MailID  string  `json:"mail_id" bson:"mail_id"`
	PhNo    string  `json:"ph_no" bson:"ph_no"`
	Address Address `json:"address" bson:"address"`
}

type Item struct {
	Name     string  `json:"name" bson:"name"`
	Price    float32 `json:"price" bson:"price"`
	Qty      int     `json:"qty" bson:"qty"`
	ImageURL string  `json:"image_url" bson:"image_url"`
	Discount int     `json:"discount" bson:"discount"`
	Total    int     `json:"total" bson:"total"`
}
type Order struct {
	ID              string          `json:"order_id" bson:"_id"`
	Customer        CustomerDetails `json:"customer" bson:"customer"`
	Items           []Item          `json:"items" bson:"items"`
	CreatedAt       string          `json:"created_at" bson:"created_at"`
	Status          string          `json:"status" bson:"status"`
	DeliveryDate    string          `json:"delivery_date" bson:"delivery_date"`
	DeliveryCharges float32         `json:"delivery_charges" bson:"delivery_charges"`
}

func (o Order) CalculateTotal() float64 {
	sum := float64(0)
	for _, item := range o.Items {
		sum += float64(item.Total)
	}
	sum += float64(o.DeliveryCharges)
	return sum
}
