package main

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string
	TelegramChatId   int64
	SearchUrls       []string
}

func InitConfig() {
	viper.SetDefault("db_path", "db.sqlite")
	viper.SetDefault("scan_interval_minutes", 10)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	for _, key := range []string{"telegram.bot_token", "telegram.chat_id", "search_urls"} {
		if !viper.IsSet(key) {
			log.Fatal("The following key is not set in the config: ", key)
		}
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name, "(Note: only `search_urls` are reloaded)")
	})
	viper.WatchConfig()
}

func Scan(dbPath string, telegramNotifier *TelegramNotifier) {
	searchUrls := viper.GetStringSlice("search_urls") // Read the search_urls from viper, so hot reloading configs is allowed
	if len(searchUrls) == 0 {
		log.Fatal("No search_urls provided or the list is empty. Please provide search urls.")
	}

	scanTime := time.Now().Unix()
	ch := make(chan []string)
	var wg sync.WaitGroup
	for _, searchUrl := range searchUrls {
		wg.Add(1)
		go (func() {
			defer wg.Done()

			listings := CrawlSearchUrl(searchUrl)

			err := WriteScanToDb(dbPath, scanTime, searchUrl, listings)
			if err != nil {
				log.Fatal(err)
			}

			urls, err := GetFirstSeenListingURLs(dbPath, scanTime, searchUrl)
			if err != nil {
				log.Fatal(err)
			}
			ch <- urls
		})()
	}

	go (func() {
		wg.Wait()
		close(ch)
	})()

	totalUrls := make([]string, 0, 10)
	for url := range ch {
		totalUrls = append(totalUrls, url...)
	}

	if len(totalUrls) > 0 {
		log.Println(len(totalUrls), " new listings:")
		for _, url := range totalUrls {
			log.Println(url)
		}
	} else {
		log.Println("No new listings")
	}

	telegramNotifier.Notify(totalUrls)
}

func main() {
	InitConfig()

	dbPath := viper.GetString("db_path")
	telegramBotToken := viper.GetString("telegram.bot_token")
	telegramChatId := viper.GetInt64("telegram.chat_id")
	sleep_duration := viper.GetDuration("scan_interval_minutes") * time.Minute

	err := Migrate(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	telegramNotifier, err := NewTelegramNotifier(telegramBotToken, telegramChatId)
	if err != nil {
		log.Fatal(err)
	}

	for {
		Scan(dbPath, telegramNotifier)

		log.Println("Sleeping", sleep_duration)
		time.Sleep(sleep_duration)
	}
}
