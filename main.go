/*
 * Copyright (C) 2015 Clinton Freeman
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"log"
	"time"
)

var mainfunc = make(chan func())

func main() {
	log.Printf("INFO: Starting scout.\n")
	var configFile string
	var videoFile string
	var debug bool

	flag.StringVar(&configFile, "configFile", "scout.json", "The path to the configuration file")
	flag.StringVar(&videoFile, "videoFile", "", "The path to a video file to detect motion from instead of a webcam")
	flag.BoolVar(&debug, "debug", false, "Should we run scout in debug mode, and render frames of detected materials")
	flag.Parse()

	config, err := parseConfiguration(configFile)
	if err != nil {
		log.Printf("INFO: %s", err)
		log.Printf("INFO: Unable to open '%s', creating one with default values.", configFile)

		// Save the default config file to disk.
		saveConfiguration(configFile, config)
	}

	// Send initial health heartbeat on startup.
	NewHeartbeat(config).post(config)

	// Send periodic health heartbeats to the mothership.
	ticker := time.NewTicker(time.Minute * 15)
	go func() {
		for range ticker.C {
			NewHeartbeat(config).post(config)
		}
	}()

	deltaC := make(chan Command)
	deltaCFG := make(chan Configuration, 1)

	go controller(deltaC, deltaCFG, configFile, config)
	monitor(deltaC, deltaCFG, videoFile, debug, config)
}
