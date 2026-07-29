package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kubotaaoto/pokedexcli/internal/pokecache"
)

func cleanInput(text string) []string {
	lower := strings.ToLower(text)
	return strings.Fields(lower)
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
}

type config struct {
	Next int
}

var cf = config{
	Next: 1,
}

var cache = pokecache.NewCache(time.Minute * 5)

type locationArea struct {
	Name string `json:"name"`
}

type Location struct {
	Name              string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int           `json:"chance"`
				ConditionValues []interface{} `json:"condition_values"`
				MaxLevel        int           `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
}

var Pokedex = map[string]Pokemon{}

var cliCommandDict = map[string]cliCommand{}

func init() {
	cliCommandDict["exit"] = cliCommand{
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	}
	cliCommandDict["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}
	cliCommandDict["map"] = cliCommand{
		name:        "map",
		description: "Displays new 20 location areas",
		callback:    commandMap,
	}
	cliCommandDict["mapb"] = cliCommand{
		name:        "mapb",
		description: "Displays the previous 20 location areas",
		callback:    commandMapb,
	}
	cliCommandDict["explore"] = cliCommand{
		name:        "explore <location_name>",
		description: "Explore a location",
		callback:    commandExplore,
	}
	cliCommandDict["catch"] = cliCommand{
		name:        "catch <pokemen_name>",
		description: "Throw a Pokeball at <pokemon>",
		callback:    commandCatch,
	}
	cliCommandDict["inspect"] = cliCommand{
		name:        "inspect <pokemen_name>",
		description: "See details about a <pokemon>",
		callback:    commandInspect,
	}
	cliCommandDict["pokedex"] = cliCommand{
		name:        "pokedex",
		description: "Displays all the pokemon the user has caught",
		callback:    commandPokedex,
	}
}

func commandExit(*config, ...string) error {
	fmt.Printf("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(*config, ...string) error {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n")
	for _, command := range cliCommandDict {
		fmt.Printf("\n%s: %s", command.name, command.description)
	}
	fmt.Printf("\n")
	return nil
}

func commandMap(cf *config, args ...string) error {
	for i := 0; i < 20; i++ {
		fullUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%d", cf.Next)
		la := locationArea{}

		if val, ok := cache.Get(fullUrl); ok {
			err := json.Unmarshal(val, &la)
			if err != nil {
				return err
			}

			cache.Add(fullUrl, val)
			return nil
		} else {
			res, err := http.Get(fullUrl)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			err = json.Unmarshal(data, &la)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			cache.Add(fullUrl, data)
		}
		fmt.Println(la.Name)
		cf.Next++
	}
	return nil
}

func commandMapb(cf *config, args ...string) error {
	cf.Next -= 40
	for i := 0; i < 20; i++ {
		fullUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%d", cf.Next)
		la := locationArea{}

		if val, ok := cache.Get(fullUrl); ok {
			err := json.Unmarshal(val, &la)
			if err != nil {
				return err
			}

			cache.Add(fullUrl, val)
			return nil
		} else {
			res, err := http.Get(fullUrl)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			err = json.Unmarshal(data, &la)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			cache.Add(fullUrl, data)
		}
		fmt.Println(la.Name)
		cf.Next++
	}
	return nil
}

func commandExplore(cfg *config, args ...string) error {
	if len(args) < 1 {
		return errors.New("you must provide a location name")
	}

	name := args[0]
	location, err := getLocation(name)
	if err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\n", location.Name)
	fmt.Println("Found Pokemon: ")
	for _, enc := range location.PokemonEncounters {
		fmt.Printf(" - %s\n", enc.Pokemon.Name)
	}
	return nil
}

func lookupCommand(name string) (cliCommand, error) {
	for key, value := range cliCommandDict {
		if key == name {
			return value, nil
		}
	}
	return cliCommand{}, fmt.Errorf("Unknown command\n")
}

func getLocation(locationName string) (Location, error) {
	url := "https://pokeapi.co/api/v2/location-area/" + locationName

	if val, ok := cache.Get(url); ok {
		locationResp := Location{}
		err := json.Unmarshal(val, &locationResp)
		if err != nil {
			return Location{}, err
		}
		cache.Add(url, val)
		return locationResp, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return Location{}, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return Location{}, err
	}

	locationResp := Location{}
	err = json.Unmarshal(data, &locationResp)
	if err != nil {
		return Location{}, err
	}
	defer res.Body.Close()
	cache.Add(url, data)
	return locationResp, nil
}

func commandCatch(cfg *config, args ...string) error {
	pokeName := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokeName)

	url := "https://pokeapi.co/api/v2/pokemon/" + pokeName
	pokemon := Pokemon{}

	if val, ok := cache.Get(url); ok {
		err := json.Unmarshal(val, &pokemon)
		if err != nil {
			return err
		}

		cache.Add(url, val)
		return nil
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &pokemon)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		cache.Add(url, data)
	}

	baseExp := pokemon.BaseExperience
	catchRate := rand.Intn(10) - baseExp/140
	if catchRate > 5 {
		Pokedex[pokeName] = pokemon
		fmt.Printf("%s was caught!\n", pokeName)
	} else {
		fmt.Printf("%s escaped!\n", pokeName)
	}
	return nil

}

func commandInspect(cfg *config, args ...string) error {
	pokeName := args[0]
	pokemon, ok := Pokedex[pokeName]
	if !ok {
		return fmt.Errorf("you have not caught that pokemon")
	}
	fmt.Print(pokemon.String())
	return nil
}

func (p Pokemon) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s\n", p.Name)
	fmt.Fprintf(&sb, "Height: %d\n", p.Height)
	fmt.Fprintf(&sb, "Weight: %d\n", p.Weight)
	sb.WriteString("Stats:\n")
	for _, stat := range p.Stats {
		fmt.Fprintf(&sb, "  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	sb.WriteString("Types:\n")
	for _, t := range p.Types {
		fmt.Fprintf(&sb, "  - %s\n", t.Type.Name)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func commandPokedex(cfg *config, args ...string) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range Pokedex {
		fmt.Printf(" - %s\n", pokemon.Name)
	}
	return nil
}
