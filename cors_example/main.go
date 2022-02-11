package main

import "net/http"

func main() {
	http.HandleFunc("/", frontHandler)
	http.ListenAndServe(":3001", nil)
}

func frontHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`	
		<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="UTF-8">
		</head>
		<body>
		<h1>Simple CORS</h1>
		<div id="output"></div>
		<script>
		document.addEventListener('DOMContentLoaded', function() {
		fetch("http://localhost:3000/v1/healthcheck").then(
		function (response) {
		response.text().then(function (text) {
		document.getElementById("output").innerHTML = text;
		});
		},
		function(err) {
		document.getElementById("output").innerHTML = err;
		}
		);
		});
		</script>
		</body>
		</html>
	`))
}
