# Torrent All-In-One

🎬 Outil CLI pour préparer des releases de films : identification TMDB (scraping), analyse technique, génération NFO et création de torrent.

## ✨ Fonctionnalités

- **Identification automatique** : Recherche le film sur TMDB via scraping (aucune clé API requise)
- **Sélection interactive** : Choix parmi les résultats ou recherche manuelle / ID direct
- **Analyse technique** : Extraction des métadonnées via MediaInfo
- **Renommage automatique** : Convention de nommage warez (Titre.Année.Résolution.Source.Codec-GROUPE)
- **Génération NFO** : Fichier NFO avec infos film et techniques
- **Présentation BBCode** : Résumé formaté pour forums
- **Création torrent** : Génération du fichier .torrent

## 🚀 Installation

### Via Docker (Recommandé)

```bash
# Cloner le repo
git clone https://github.com/metwurcht/torrent-all-in-one.git
cd torrent-all-in-one

# Installer
chmod +x install.sh
./install.sh
```

### Compilation manuelle

```bash
# Prérequis: Go 1.21+, mediainfo

go mod download
go build -o torrent-aio ./cmd/torrent-aio
```

## 📖 Utilisation

### Commande de base

```bash
torrent-aio /chemin/vers/film.mkv
```

### Options

```bash
torrent-aio film.mkv \
  --tracker "http://tracker.example.com/announce" \
  --group "MONGROUPE" \
  --output /chemin/sortie
  --no-rename          # Ne pas renommer le fichier
  --skip-torrent      # Ne pas générer le fichier torrent
```

### Fichier de configuration

Créez `~/.config/torrent-aio.yml` :

```yaml
group_name: "MONGROUPE"
```

## 🔧 Workflow

1. **Analyse parallèle** : Le fichier est analysé en arrière-plan pendant la recherche TMDB
2. **Recherche TMDB** : Les mots-clés sont extraits du nom de fichier (scraping web)
3. **Sélection** : Choisissez le bon film dans la liste ou :
   - Tapez `0` pour une nouvelle recherche
   - Entrez `id:12345` pour utiliser un ID TMDB directement
4. **Génération** :
   - Le fichier est renommé selon la convention warez
   - Un fichier NFO est créé
   - Le résumé bbcode est affiché dans la console
   - Le fichier torrent est généré

## 🏗️ Architecture

```
torrent-all-in-one/
├── cmd/torrent-aio/      # Point d'entrée CLI
├── internal/
│   ├── cli/              # Commandes Cobra
│   ├── tmdb/             # Client TMDB (scraping web)
│   ├── mediainfo/        # Analyse fichiers vidéo
│   ├── nfo/              # Génération NFO
│   ├── renamer/          # Renommage warez
│   ├── presenter/        # Génération présentation BBCode
│   ├── torrent/          # Génération torrent
│   └── ui/               # Interface utilisateur
├── scripts/              # Scripts wrapper Docker
└── Dockerfile
```

## 🔌 Intégration

L'architecture modulaire permet une intégration facile :

### Comme bibliothèque Go

```go
import (
    "github.com/metwurcht/torrent-all-in-one/internal/tmdb"
    "github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
)

// Client TMDB (scraping, aucune clé API nécessaire)
client := tmdb.NewClient()
movie, _ := client.GetMovieDetails(ctx, 12345)

// Analyse fichier
analyzer := mediainfo.NewAnalyzer()
info, _ := analyzer.Analyze("/path/to/file.mkv")
```

### Via API REST (à venir)

Le package `ui.Prompter` permet de remplacer l'interface CLI par une API :

```go
// Utiliser le SilentPrompter pour l'automatisation
prompter := ui.NewSilentPrompter()
prompter.SetDefaultMovieIndex(0) // Sélection auto du premier résultat
```

## 🐳 Docker

### Build manuel

```bash
docker build -t torrent-aio:latest .
```

### Utilisation directe

**Important** : Utilisez toujours `-it` pour l'interface interactive

```bash
docker run --rm -it \
  -v /chemin/local:/data \
  torrent-aio:latest /data/film.mkv
```

> ⚠️ Sans `-it`, l'application ne pourra pas lire votre entrée (erreur EOF)

### Docker Compose

```bash
docker-compose run torrent-aio /data/film.mkv
```

## 📋 Prérequis

- **Docker** (recommandé) ou
- **Go 1.21+** pour la compilation
- **MediaInfo** pour l'analyse des fichiers

## 📝 Licence

MIT

## 🤝 Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une PR.
