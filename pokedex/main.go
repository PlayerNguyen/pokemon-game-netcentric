package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
)

const API_GET_ALL_POKEMONS_URL = "https://pokeapi.co/api/v2/pokemon?limit=10000"

type FetchPokemonResult struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type FetchPokemon struct {
	Result []FetchPokemonResult `json:"results"`
}

// type MonsterStats struct {
// 	Hp             int `json:"hp"`
// 	Attack         int `json:"attack"`
// 	Defense        int `json:"defense"`
// 	SpecialAttack  int `json:"special_attack"`
// 	SpecialDefense int `json:"special_defense"`
// 	Speed          int `json:"speed"`
// }

type MonsterStatsProperty struct {
	Name string `json:"name"`
}
type MonsterStatsImport struct {
	BaseStats int                  `json:"base_stat"`
	Effort    int                  `json:"effort"`
	Stat      MonsterStatsProperty `json:"stat"`
}

type Monster struct {
	Name           string               `json:"name"`
	Id             int                  `json:"id"`
	BaseExperience int                  `json:"base_experience"`
	Order          int                  `json:"order"`
	Stats          []MonsterStatsImport `json:"stats"`
}

type MonsterSafeStorage struct {
	mutex    sync.Mutex
	monsters []Monster
}

func PanicIfError(err error) {
	if err != nil {
		log.Panicln(err.Error())
	}
}

func fetch_all_pokemons() FetchPokemon {

	res, err := http.Get(API_GET_ALL_POKEMONS_URL)
	PanicIfError(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	PanicIfError(err)

	fetch_pokemon := FetchPokemon{}
	json.Unmarshal(body, &fetch_pokemon)

	return fetch_pokemon
}

func fetch_monster_and_store(result FetchPokemonResult, idx int, wg *sync.WaitGroup, storage *MonsterSafeStorage) {
	log.Printf("Fetching pokemon in url %s\n", result.Url)

	res, err := http.Get(result.Url)
	PanicIfError(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	PanicIfError(err)

	monster := Monster{}
	json.Unmarshal(body, &monster)

	// log.Println()
	output_data, err := json.Marshal((monster))
	PanicIfError(err)

	os.Mkdir("monsters", os.ModePerm)

	_err := os.WriteFile(path.Join("monsters", strconv.Itoa(idx)+".json"), output_data, 0755)
	PanicIfError(_err)

	wg.Done()

	// append into a storage
	storage.mutex.Lock()
	storage.monsters = append(storage.monsters, monster)
	storage.mutex.Unlock()
}

func write_pokedex_json(storage *MonsterSafeStorage) {
	// Sort the storage
	storage.mutex.Lock()
	sort.Slice(storage.monsters, func(i, j int) bool {
		return storage.monsters[i].Id < storage.monsters[j].Id
	})
	log.Println("Write pokedex.json file")
	b, err := json.Marshal(storage.monsters)
	PanicIfError(err)
	_err := os.WriteFile("pokedex.json", b, 0755)
	PanicIfError(_err)
	storage.mutex.Unlock()
}

func main() {
	wg := sync.WaitGroup{}
	storage := MonsterSafeStorage{}
	pokemons := fetch_all_pokemons()
	for idx, pokemon := range pokemons.Result {
		wg.Add(1)
		go fetch_monster_and_store(pokemon, idx, &wg, &storage)
	}

	wg.Wait()

	// After finish fetching, convert into a pokedex.json file
	write_pokedex_json(&storage)
}
