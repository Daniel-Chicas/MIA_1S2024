package main

import (
	"MIA_Proyecto2_201807079/Comandos"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var logued = false

type DatosEntrada struct {
	Comandos []string `json:"comandos"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", inicial).Methods("GET")
	router.HandleFunc("/analizador", analizador).Methods("POST")

	handler := allowCORS(router)
	fmt.Println("Server on port :3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

func allowCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		handler.ServeHTTP(w, r)
	})
}

func inicial(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>¡Hola Desde el servidor!</h1>")
}

func analizador(w http.ResponseWriter, r *http.Request) {
	var datos DatosEntrada
	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = guardarDatos("./prueba.script", datos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ejecutar el archivo de script
	Exec("./prueba.script")
	fmt.Fprintf(w, "Script ejecutado exitosamente")
}

func guardarDatos(archivo string, datos DatosEntrada) error {
	// Abrir o crear el archivo
	file, err := os.Create(archivo)
	if err != nil {
		return err
	}
	defer file.Close()

	// Escribir los comandos en el archivo
	for _, comando := range datos.Comandos {
		_, err := file.WriteString(strings.TrimSpace(comando) + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func Comando(text string) string {
	var tkn string
	terminar := false
	for i := 0; i < len(text); i++ {
		if terminar {
			if string(text[i]) == " " || string(text[i]) == "-" {
				break
			}
			tkn += string(text[i])
		} else if string(text[i]) != " " && !terminar {
			if string(text[i]) == "#" {
				tkn = text
			} else {
				tkn += string(text[i])
				terminar = true
			}
		}
	}
	return tkn
}

func SepararTokens(texto string) []string {
	var tokens []string
	if texto == "" {
		return tokens
	}
	texto += " "
	var token string
	estado := 0
	for i := 0; i < len(texto); i++ {
		c := string(texto[i])
		if estado == 0 && c == "-" {
			estado = 1
		} else if estado == 0 && c == "#" {
			continue
		} else if estado != 0 {
			if estado == 1 {
				if c == "=" {
					estado = 2
				} else if c == " " {
					continue
				} else if (c == "P" || c == "p") && string(texto[i+1]) == " " && string(texto[i-1]) == "-" {
					estado = 0
					tokens = append(tokens, c)
					token = ""
					continue
				} else if (c == "R" || c == "r") && string(texto[i+1]) == " " && string(texto[i-1]) == "-" {
					estado = 0
					tokens = append(tokens, c)
					token = ""
					continue
				}
			} else if estado == 2 {
				if c == " " {
					continue
				}
				if c == "\"" {
					estado = 3
					continue
				} else {
					estado = 4
				}
			} else if estado == 3 {
				if c == "\"" {
					estado = 4
					continue
				}
			} else if estado == 4 && c == "\"" {
				tokens = []string{}
				continue
			} else if estado == 4 && c == " " {
				estado = 0
				tokens = append(tokens, token)
				token = ""
				continue
			}
			token += c
		}
	}
	return tokens
}

func funciones(token string, tks []string) {
	if token != "" {
		if Comandos.Comparar(token, "MKDISK") {
			fmt.Println("*************************************** FUNCIÓN MKDISK **************************************")
			Comandos.ValidarDatosMKDISK(tks)
		} else if Comandos.Comparar(token, "RMDISK") {
			fmt.Println("*************************************** FUNCIÓN RMDISK **************************************")
			Comandos.RMDISK(tks)
		} else if Comandos.Comparar(token, "FDISK") {
			fmt.Println("*************************************** FUNCIÓN FDISK  **************************************")
			Comandos.ValidarDatosFDISK(tks)
		} else if Comandos.Comparar(token, "MOUNT") {
			fmt.Println("*************************************** FUNCIÓN MOUNT  **************************************")
			Comandos.ValidarDatosMOUNT(tks)
		} else if Comandos.Comparar(token, "MKFS") {
			fmt.Println("*************************************** FUNCIÓN MKFS  **************************************")
			Comandos.ValidarDatosMKFS(tks)
		} else if Comandos.Comparar(token, "REP") {
			fmt.Println("*************************************** FUNCIÓN REP  **************************************")
			Comandos.ValidarDatosREP(tks)
		} else if Comandos.Comparar(token, "LOGIN") {
			fmt.Println("*************************************** FUNCIÓN LOGIN  **************************************")
			if logued {
				Comandos.Error("LOGIN", "Ya hay un usuario en línea.")
				return
			} else {
				logued = Comandos.ValidarDatosLOGIN(tks)
			}
		} else if Comandos.Comparar(token, "LOGOUT") {
			fmt.Println("*************************************** FUNCIÓN LOGOUT  **************************************")
			if !logued {
				Comandos.Error("LOGOUT", "Aún no se ha iniciado sesión.")
				return
			} else {
				logued = Comandos.CerrarSesion()
			}
		} else if Comandos.Comparar(token, "MKGRP") {
			fmt.Println("*************************************** FUNCIÓN MKGRP  **************************************")
			if !logued {
				Comandos.Error("MKGRP", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosGrupos(tks, "MK")
			}
		} else if Comandos.Comparar(token, "RMGRP") {
			fmt.Println("*************************************** FUNCIÓN RMGRP  **************************************")
			if !logued {
				Comandos.Error("RMGRP", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosGrupos(tks, "RM")
			}
		} else if Comandos.Comparar(token, "MKUSER") {
			fmt.Println("*************************************** FUNCIÓN MKUSER  **************************************")
			if !logued {
				Comandos.Error("MKUSER", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "MK")
			}
		} else if Comandos.Comparar(token, "RMUSR") {
			fmt.Println("*************************************** FUNCIÓN RMUSER  **************************************")
			if !logued {
				Comandos.Error("RMUSER", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "RM")
			}
		} else if Comandos.Comparar(token, "MKDIR") {
			fmt.Println("*************************************** FUNCIÓN MKDIR  **************************************")
			if !logued {
				Comandos.Error("MKDIR", "Aún no se ha iniciado sesión.")
				return
			} else {
				var p string
				particion := Comandos.GetMount("MKDIR", Comandos.Logged.Id, &p)
				Comandos.ValidarDatosMKDIR(tks, particion, p)
			}
		} else if Comandos.Comparar(token, "MKFILE") {
			fmt.Println("*************************************** FUNCIÓN MKFILE  **************************************")
			if !logued {
				Comandos.Error("MKFILE", "Aún no se ha iniciado sesión.")
				return
			} else {
				var p string
				particion := Comandos.GetMount("MKDIR", Comandos.Logged.Id, &p)
				Comandos.ValidarDatosMKFILE(tks, particion, p)
			}
		} else {
			Comandos.Error("ANALIZADOR", "No se reconoce el comando \""+token+"\"")
		}
	}
}

func Exec(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error al abrir el archivo: %s", err)
	}
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		texto := fileScanner.Text()
		texto = strings.TrimSpace(texto)
		tk := Comando(texto)
		if texto != "" {
			if Comandos.Comparar(tk, "pause") {
				fmt.Println("************************************** FUNCIÓN PAUSE **************************************")
				var pause string
				Comandos.Mensaje("PAUSE", "Presione \"enter\" para continuar...")
				fmt.Scanln(&pause)
				continue
			} else if string(texto[0]) == "#" {
				fmt.Println("************************************** COMENTARIO **************************************")
				Comandos.Mensaje("COMENTARIO", texto)
				continue
			}
			texto = strings.TrimLeft(texto, tk)
			tokens := SepararTokens(texto)
			funciones(tk, tokens)
		}
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error al leer el archivo: %s", err)
	}
	file.Close()
}
