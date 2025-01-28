package parser

import (
	"strconv"
	"testing"
)

var programs = []string{
	`Declare String username, address, city, state, phone, major
Declare Integer zip
Display "What is your name?"
Input username
Display "What is your street address?"
Input address
Display "What city do you live in?"
Input city
Display "What state do you live in?"
Input state
Display "What’s your zipcode?"
Input zip
Display "What’s your phone number?"
Input phone
Display "What’s your college major?"
Input major
Display "Name: ", username
Display "Address:"
Display address
Display city, ", ", state, " ", zip
Display "Phone number: ", phone
Display "College major: ", major
`, `Constant Real PROFIT_MARGIN = 0.23
Declare Integer sales
Display "What were the annual sales?"
Input sales
Display "The company made about $", sales * PROFIT_MARGIN
`, `Constant Real LAND_DIVISOR = 43560
Declare Real sqrFeet
Display "How many square feet of land are in this tract?"
Input sqrFeet
Display "This tract takes up ", sqrFeet / LAND_DIVISOR, " acres of land."
`, `Constant Real TAX = 0.06
Declare Real itemOnePrice, itemTwoPrice, itemThreePrice, itemFourPrice, itemFivePrice
Display "How much does a bag of pecans cost?"
Input itemOnePrice
Display "How much does a bar of soap cost?"
Input itemTwoPrice
Display "How much does a bag of chips cost?"
Input itemThreePrice
Display "How much does a bottle of shampoo cost?"
Input itemFourPrice
Display "How much does a bottle of laundry detergent cost?"
Input itemFivePrice
Declare Real subtotal = itemOnePrice + itemTwoPrice + itemThreePrice + itemFourPrice + itemFivePrice
Display "Subtotal: $", subtotal
Display "Tax: $", subtotal * TAX
Display "Your grand total is $", subtotal + (subtotal * TAX)
`, `Constant Real MILES_PER_HOUR = 60
Display "In 5 hours, the car will travel ", MILES_PER_HOUR * 5, " miles."
Display "In 8 hours, the car will travel ", MILES_PER_HOUR * 8, " miles."
Display "In 12 hours, the car will travel ", MILES_PER_HOUR * 12, " miles."
`, `Constant Real STATE_TAX_RATE = 0.04
Constant Real COUNTY_TAX_RATE = 0.02
Declare Real subtotal, stateTax, countyTax
Display "What is the original amount of the purchase?"
Input subtotal
Set stateTax = subtotal * STATE_TAX_RATE
Set countyTax = subtotal * COUNTY_TAX_RATE
Display "Subtotal: $", subtotal
Display "State tax: $", stateTax
Display "County tax: $", countyTax
Display "Total tax: $", stateTax + countyTax
Display "Total sale: $", subtotal + stateTax + countyTax
`, `Declare Real milesDriven, gallonsUsed, milesPerGallon
Display "How many miles did you drive?"
Input milesDriven
Display "How many gallons of gas did you use?"
Input gallonsUsed
Set milesPerGallon = milesDriven / gallonsUsed
Display "Miles per gallon: ", milesPerGallon
`, `Constant Real TIP_RATE = 0.15
Constant Real SALES_TAX_RATE = 0.04
Declare Real subtotal, tip, tax
Display "How much does the meal cost?"
Input subtotal
Set tip = subtotal * TIP_RATE
Set tax = subtotal * SALES_TAX_RATE
Display "Subtotal: $", subtotal
Display "Tip: $", tip
Display "Tax: $", tax
Display "Total: $", subtotal + tip + tax
`, `Constant Real WEIGHT_CHANGE = 4
Declare Real startingWeight
Display "What is your starting weight?"
Input startingWeight
Display "After one months of reducing calories by 500, your weight will be ", startingWeight - (1 * WEIGHT_CHANGE), "."
Display "After two months of reducing calories by 500, your weight will be ", startingWeight - (2 * WEIGHT_CHANGE), "."
Display "After three months of reducing calories by 500, your weight will be ", startingWeight - (3 * WEIGHT_CHANGE), "."
Display "After four months of reducing calories by 500, your weight will be ", startingWeight - (4 * WEIGHT_CHANGE), "."
Display "After five months of reducing calories by 500, your weight will be ", startingWeight - (5 * WEIGHT_CHANGE), "."
Display "After six months of reducing calories by 500, your weight will be ", startingWeight - (6 * WEIGHT_CHANGE), "."
`, `Declare Real monthlyPayment
Declare Integer monthsPaid
Display "What’s your monthly car payment?"
Input monthlyPayment
Display "How many months have you paid up?"
Input monthsPaid
Display "Overall, you’ve paid $", monthlyPayment * monthsPaid, " on car payments."
`, `Declare Integer wholePizzas, slicesPerPizza, people, leftOverSlices
Display "How many whole pizzas are there?"
Input wholePizzas
Display "How many slices is each pizza cut into?"
Input slicesPerPizza
Display "Each person will get 3 slices. How many people are there?"
Input people
Set leftOverSlices = (wholePizzas * slicesPerPizza) - (3 * people)
Display "There will be ", leftOverSlices, " slices left over."
`, `Declare Real celsius, fahrenheit
Display "How many degrees Celsius is it outside?"
Input celsius
Set fahrenheit = (1.8 * celsius) + 32
Display "It’s ", fahrenheit, " degrees fahrenheit outside."
`,
}

func TestPrograms(t *testing.T) {
	for i, input := range programs {
		id := strconv.Itoa(i + 1)
		t.Run(id, func(t *testing.T) {
			block, err := Parse([]byte(input))
			if block != nil {
				for _, stmt := range block.Statements {
					t.Log(stmt.String())
				}
			}
			if err != nil {
				t.Error(err)
			}
		})
	}
}
