package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// This program generates a log file which get rotated automatically,
// and you can have another process of Sentlog consuming this log file.
func main() {
	logger := &lumberjack.Logger{
		Filename:   "ignore_me.log",
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	defer logger.Close()

	go func() {
		// This goroutine handles automatic rotate per 1 minute
		// You can change this to whatever you like.
		for {
			time.Sleep(time.Minute)
			err := logger.Rotate()
			if err != nil {
				log.Error().Err(err).Msg("rotating logger")
			}
		}
	}()

	var texts = []string{
		"Lorem ipsum dolor sit amet, consectetuer adipiscing elit",
		"Aenean commodo ligula eget dolor",
		"Aenean massa",
		"Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus",
		"Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem",
		"Nulla consequat massa quis enim",
		"Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu",
		"In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo",
		"Nullam dictum felis eu pede mollis pretium",
		"Integer tincidunt",
		"Cras dapibus",
		"Vivamus elementum semper nisi",
		"Aenean vulputate eleifend tellus",
		"Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim",
		"Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus",
		"Phasellus viverra nulla ut metus varius laoreet",
		"Quisque rutrum",
		"Aenean imperdiet",
		"Etiam ultricies nisi vel augue",
		"Curabitur ullamcorper ultricies nisi",
		"Nam eget dui",
		"Etiam rhoncus",
		"Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum",
		"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem",
		"Maecenas nec odio et ante tincidunt tempus",
		"Donec vitae sapien ut libero venenatis faucibus",
		"Nullam quis ante",
		"Etiam sit amet orci eget eros faucibus tincidunt",
		"Duis leo",
		"Sed fringilla mauris sit amet nibh",
		"Donec sodales sagittis magna",
		"Sed consequat, leo eget bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero",
		"Fusce vulputate eleifend sapien",
		"Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id, metus",
		"Nullam accumsan lorem in dui",
		"Cras ultricies mi eu turpis hendrerit fringilla",
		"Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; In ac dui quis mi consectetuer lacinia",
		"Nam pretium turpis et arcu",
		"Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum",
		"Sed aliquam ultrices mauris",
		"Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris",
		"Praesent adipiscing",
		"Phasellus ullamcorper ipsum rutrum nunc",
		"Nunc nonummy metus",
		"Vestibulum volutpat pretium libero.",
	}

	for {
		var randomIndex = rand.Intn(len(texts) - 1)

		entry := fmt.Sprintf("[%s] n=%d - %s\n", time.Now().Format(time.RubyDate), randomIndex, texts[randomIndex])
		_, err := logger.Write([]byte(entry))
		if err != nil {
			log.Warn().Err(err).Msg("writing entry to log file")
		}

		time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
	}
}
