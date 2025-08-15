# WasmP2PChat 🛰️

Une application de chat utilisant le canal de données P2P de WebRTC, implémentée en Go WebAssembly, permettant de trouver et de discuter avec des inconnus au hasard.

## Résumé ✴️

- **Objectif** ：Démonstration pour apprendre les bases de l’intégration de Wasm et WebRTC
- **Fonctionnalités** ：
  - Appairage aléatoire pour une connexion P2P
  - Envoi et réception de messages texte
- **Public visé** ：Ingénieurs souhaitant tester WebRTC et Wasm (niveau intermédiaire)

## Utilisation 🧑‍💻
### Prérequis

- Go 1.24.x（compatible Wasm）
- Navigateur moderne (Chrome / Firefox / Safari)
- 3 fenêtres de terminal
- Recommandé pour une utilisation en réseau local (personnalisation nécessaire pour un usage en production)

## Fonctionnement du chat 🧐

1. Compiler le fichier WebAssembly
```bash
GOOS=js GOARCH=wasm go build -o main.wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./wasm_exec.js
```

2. Terminal 1 : Lancer le serveur de fichiers statiques
```bash
python3 -m http.server 8080
```

3. Terminal 2 : Lancer le serveur de matchmaking
```bash
go run cmd/matchmaking/main.go
```

4. Terminal 3 : Lancer le serveur de signalisation
```bash
go run cmd/signaling/main.go
```

5. Ouvrir dans le navigateur
- Ouvrir http://localhost:8080 dans deux onglets distincts
- Appuyer sur START dans les deux onglets
- Saisir un message et appuyer sur SEND pour échanger des messages

## Pile technologique 🛠️

- WebAssembly (Wasm)
  - Format binaire permettant d’exécuter des langages comme Go dans le navigateur
  - Performances élevées, typage sûr, et support de la concurrence de Go
- WebRTC
  - Technologie de communication en temps réel pour des connexions directes entre navigateurs
  - Supporte audio, vidéo et données
- DataChannel
  - Fonction de communication de données de WebRTC
  - Permet d’envoyer du texte ou des données binaires directement
- Communication P2P
  - Communication directe entre appareils sans serveur intermédiaire
  - Réduit la charge serveur et minimise la latence

## Avantages et limites 👩‍⚕️

```bash
[Browser A] --WebSocket--> [Serveur de matchmaking] <--WebSocket-- [Browser B]
       ↓ Notification RoomID                      ↓
[Browser A] --WebSocket--> [Serveur de signalisation] <--WebSocket-- [Browser B]
       ↓ Échange d’informations de connexion (SDP, ICE candidates)
[Browser A] <==== WebRTC DataChannel (P2P) ====> [Browser B]
```

1. Matchmaking
    - Met en relation deux utilisateurs et attribue un identifiant de salle (Room ID)

2. Signalisation
    - Échange les SDP et les candidats ICE nécessaires pour la connexion WebRTC

3. Communication P2P
    - Une fois connectés, les navigateurs échangent les données directement

## Dépannage ⛓️‍💥

| Symptôme                            | Cause                           | Solution                |
| ----------------------------------- | ------------------------------- | ----------------------- |
| Impossible de se connecter          | Serveur de signalisation arrêté | Vérifier le terminal 3  |
| Aucun pair trouvé                   | Serveur de matchmaking arrêté   | Vérifier le terminal 2  |
| Ne fonctionne pas hors réseau local | Échec de traversée NAT          | Ajouter un serveur TURN |


## Références 🔖

- [https://webrtc.org/?hl=ja](https://webrtc.org/?hl=ja)
- [https://developer.mozilla.org/ja/docs/WebAssembly/Guides/Concepts](https://developer.mozilla.org/ja/docs/WebAssembly/Guides/Concepts)
- [https://github.com/OpenAyame/ayame](https://github.com/OpenAyame/ayame)