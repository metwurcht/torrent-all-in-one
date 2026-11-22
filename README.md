# Torrent All-In-One

🎬 Outil CLI pour préparer des releases de films : identification TMDB via l'API officielle, analyse technique, génération NFO et création de torrent.

> ⚠️ **Prérequis important** : [MediaInfo](https://mediaarea.net/en/MediaInfo) doit être installé sur votre système pour que l'outil fonctionne correctement.

## ✨ Fonctionnalités

- **Identification automatique** : Recherche le film ou la série via l'API officielle TMDB
- **Sélection interactive** : Choix parmi les résultats ou recherche manuelle / ID direct
- **Analyse technique** : Extraction des métadonnées via MediaInfo
- **Renommage automatique** : Convention de nommage warez (Titre.Année.Résolution.Source.Codec-GROUPE)
- **Génération NFO** : Fichier NFO avec infos film et techniques
- **Présentation BBCode** : Résumé formaté pour forums
- **Création torrent** : Génération du fichier .torrent

## 🚀 Installation

### ⚠️ Prérequis : MediaInfo

Avant d'installer Torrent All-In-One, assurez-vous d'avoir MediaInfo installé :

**Linux (Debian/Ubuntu)** :

```bash
sudo apt install mediainfo
```

**Linux (Fedora)** :

```bash
sudo dnf install mediainfo
```

**macOS** :

```bash
brew install media-info
```

**Windows** :

```powershell
winget install MediaArea.MediaInfo
```

Ou téléchargez depuis [mediaarea.net](https://mediaarea.net/en/MediaInfo/Download)

---

## 🖥️ Interface en Ligne de Commande (CLI)

### Installation du CLI

#### Option 1 : Installation automatique (recommandé)

**Linux/macOS** :

```bash
curl -sSL https://raw.githubusercontent.com/metwurcht/torrent-all-in-one/main/install.sh | bash
```

Ou avec wget :

```bash
wget -qO- https://raw.githubusercontent.com/metwurcht/torrent-all-in-one/main/install.sh | bash
```

**Windows** :

```powershell
irm https://raw.githubusercontent.com/metwurcht/torrent-all-in-one/main/install.ps1 | iex
```

Ou téléchargez et exécutez [install.ps1](install.ps1) :

```powershell
powershell -ExecutionPolicy Bypass -File install.ps1
```

#### Option 2 : Installation manuelle depuis les releases

1. Téléchargez le binaire correspondant à votre système depuis les [releases GitHub](https://github.com/metwurcht/torrent-all-in-one/releases) :

**Linux** :

- `torrent-aio-linux-amd64` - Linux x86_64
- `torrent-aio-linux-arm64` - Linux ARM64

**macOS** :

- `torrent-aio-darwin-amd64` - macOS Intel
- `torrent-aio-darwin-arm64` - macOS Apple Silicon

**Windows** :

- `torrent-aio-windows-amd64.exe` - Windows 64-bit
- `torrent-aio-windows-arm64.exe` - Windows ARM64

2. Installez le binaire :

**Linux/macOS** :

```bash
# Télécharger (remplacez VERSION et PLATFORM par les valeurs appropriées)
curl -L -o torrent-aio https://github.com/metwurcht/torrent-all-in-one/releases/latest/download/torrent-aio-linux-amd64

# Rendre exécutable
chmod +x torrent-aio

# Installer dans le PATH
sudo mv torrent-aio /usr/local/bin/
```

**Windows** :

- Téléchargez le `.exe` depuis les releases
- Placez-le dans un dossier de votre choix
- Ajoutez ce dossier à votre PATH

### Utilisation du CLI

#### Commande de base

```bash
torrent-aio /chemin/vers/film.mkv     # Film
torrent-aio /chemin/vers/serie/       # Série TV (dossier d'épisodes)
torrent-aio /chemin/vers/album/       # Album (dossier de fichiers audio)
torrent-aio --help                    # Aide
torrent-aio --version                 # Version
```

Au premier lancement, si une clé API nécessaire est introuvable, l'outil s'arrête
et indique la commande à exécuter pour la configurer.

#### Options avancées

```bash
torrent-aio film.mkv \
  --group "MONGROUPE" \
  --output /chemin/sortie \
  --no-rename          # Ne pas renommer le fichier
  --skip-torrent       # Ne pas générer le fichier torrent
```

#### Configuration

Les paramètres (clés API, nom de groupe, etc.) se gèrent avec la commande `config` :

```bash
torrent-aio config set tmdb_api_key <votre_clé>
torrent-aio config set group_name MONGROUPE
torrent-aio config list      # Valeurs effectives et leur provenance
torrent-aio config --help    # Liste des paramètres disponibles
```

Chaque paramètre est lu dans cet ordre : flag de la ligne de commande, fichier de
configuration (`~/.config/torrent-aio.yml`, sinon `~/torrent-aio.yml`), variable
d'environnement, valeur par défaut.

#### Mise à jour

À chaque exécution, l'outil vérifie si une nouvelle version est disponible sur GitHub
et l'indique en fin d'exécution. Pour l'installer :

```bash
torrent-aio update          # sudo torrent-aio update si installé dans /usr/local/bin
```

La vérification peut être désactivée avec `torrent-aio config set check_updates false`.

---

## 🔧 Workflow

1. **Analyse parallèle** : Le fichier est analysé en arrière-plan pendant la recherche TMDB
2. **Recherche TMDB** : Les mots-clés sont extraits du nom de fichier puis envoyés à l'API officielle
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
├── cmd/
│   └── torrent-aio/      # Point d'entrée en ligne de commande
├── internal/
│   ├── cli/              # Commandes Cobra
│   ├── tmdb/             # Client de l'API officielle TMDB
│   ├── mediainfo/        # Analyse fichiers vidéo
│   ├── nfo/              # Génération NFO
│   ├── renamer/          # Renommage warez
│   ├── presenter/        # Génération présentation BBCode
│   ├── torrent/          # Génération torrent
│   ├── update/           # Vérification et installation des mises à jour
│   └── ui/               # Interaction utilisateur (prompts)
```

## 📝 Licence

MIT

## 🤝 Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une PR.
