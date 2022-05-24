package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ResponseStruct struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []struct {
		PairString      string `json:"pair_string"`
		Price           string `json:"price"`
		Vol             string `json:"vol"`
		PriceChange24Hr string `json:"price_change24hr"`
		MarketCap       string `json:"market_cap"`
		High            string `json:"high"`
		Low             string `json:"low"`
	} `json:"data"`
}

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Could not load environment file.")
	}

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))

	go printCommandEvents(bot.CommandEvents())

	bot.Command("ping", &slacker.CommandDefinition{
		Handler: func(botContext slacker.BotContext, request slacker.Request, writer slacker.ResponseWriter) {
			err := writer.Reply("pong")
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	requestBody, err := json.Marshal(map[string]string{
		"pair_string": "btc_usdt",
	})

	if err != nil {
		log.Fatal(err)
	}

	bot.Command("btc_usdt", &slacker.CommandDefinition{
		Handler: func(botContext slacker.BotContext, request slacker.Request, writer slacker.ResponseWriter) {
			//err := writer.Reply("pong")
			//if err != nil {
			//	log.Fatal(err)
			//}
			resp, err := http.Post("http://13.232.254.157/api/get_prices", "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			var objmap ResponseStruct
			err = json.Unmarshal(body, &objmap)
			if err != nil {
				log.Fatal(err)
			}

			err = writer.Reply(fmt.Sprintf("PairString: %s, \tPrice: %s,  \tVolume: %s,\tPriceChange24Hour: %s, "+
				"\tMarketCap: %s, \tHigh: %s, \tLow: %s",
				objmap.Data[0].PairString, objmap.Data[0].Price, objmap.Data[0].Vol, objmap.Data[0].PriceChange24Hr,
				objmap.Data[0].MarketCap, objmap.Data[0].High, objmap.Data[0].Low))

			//bodyString, err := json.Marshal(objmap.Data[0])
			//
			//err = writer.Reply(string(bodyString))
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
