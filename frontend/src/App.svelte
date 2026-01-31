<script>
  import { onMount } from "svelte";
  import {
    ProcessFile,
    SelectFile,
    SelectDirectory,
    GetDefaultConfig,
    SelectMovie,
    SelectSourceType,
    RespondConfirm,
  } from "../wailsjs/go/gui/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";

  let filePath = "";
  let outputDir = "";
  let groupName = "TORRENT-AIO";
  let skipTorrent = false;
  let noRename = false;
  let processing = false;
  let progressLogs = [];
  let result = null;

  // Sélections
  let movieSelectionPending = false;
  let movieOptions = [];
  let sourceTypeSelectionPending = false;
  let sourceTypeOptions = [];
  let confirmPending = false;
  let confirmMessage = "";

  onMount(async () => {
    // Charger la configuration par défaut
    const config = await GetDefaultConfig();
    groupName = config.groupName;
    skipTorrent = config.skipTorrent;
    noRename = config.noRename;

    // Écouter les événements de progression
    EventsOn("progress:start", (message) => {
      addLog(message);
    });

    EventsOn("progress:update", (message) => {
      addLog(message);
    });

    EventsOn("progress:complete", (message) => {
      addLog(message);
    });

    EventsOn("progress:error", (message) => {
      addLog("❌ " + message);
    });

    // Écouter les demandes de sélection
    EventsOn("movie-selection-request", (movies) => {
      movieOptions = movies;
      movieSelectionPending = true;
    });

    EventsOn("source-type-selection-request", (sourceTypes) => {
      sourceTypeOptions = sourceTypes;
      sourceTypeSelectionPending = true;
    });

    EventsOn("confirm-request", (message) => {
      confirmMessage = message;
      confirmPending = true;
    });
  });

  function addLog(message) {
    progressLogs = [...progressLogs, message];
    // Auto-scroll vers le bas
    setTimeout(() => {
      const logElement = document.querySelector(".progress-log");
      if (logElement) {
        logElement.scrollTop = logElement.scrollHeight;
      }
    }, 10);
  }

  async function handleSelectFile() {
    filePath = await SelectFile();
  }

  async function handleSelectDirectory() {
    outputDir = await SelectDirectory();
  }

  async function handleProcess() {
    if (!filePath) {
      alert("Veuillez sélectionner un fichier");
      return;
    }

    processing = true;
    progressLogs = [];
    result = null;

    try {
      const response = await ProcessFile({
        filePath,
        outputDir,
        groupName,
        skipTorrent,
        noRename,
        sourceType: null,
      });

      if (response.success) {
        result = response;
        addLog("");
        addLog("✅ Traitement terminé avec succès !");
      } else {
        addLog("❌ Erreur: " + response.error);
      }
    } catch (error) {
      addLog("❌ Erreur: " + error.message);
    } finally {
      processing = false;
    }
  }

  function clearLogs() {
    progressLogs = [];
    result = null;
  }

  async function handleMovieSelection(movieID) {
    await SelectMovie(movieID);
    movieSelectionPending = false;
    movieOptions = [];
  }

  async function handleSourceTypeSelection(sourceType) {
    await SelectSourceType(sourceType);
    sourceTypeSelectionPending = false;
    sourceTypeOptions = [];
  }

  async function handleConfirmResponse(response) {
    await RespondConfirm(response);
    confirmPending = false;
    confirmMessage = "";
  }
</script>

<main>
  <h1>🎬 Torrent All-In-One</h1>
  <p class="subtitle">Préparation de releases de films</p>

  <!-- Modal de sélection de film -->
  {#if movieSelectionPending}
    <div class="modal-overlay">
      <div class="modal">
        <h2>Sélectionnez le film</h2>
        <div class="movie-list">
          {#each movieOptions as movie}
            <button
              class="movie-card"
              on:click={() => handleMovieSelection(movie.id)}
            >
              {#if movie.posterPath}
                <img src={movie.posterPath} alt={movie.title} />
              {/if}
              <div class="movie-info">
                <h3>{movie.title}</h3>
                {#if movie.originalTitle !== movie.title}
                  <p class="original-title">{movie.originalTitle}</p>
                {/if}
                <p class="release-date">{movie.releaseDate}</p>
                {#if movie.overview}
                  <p class="overview">{movie.overview}</p>
                {/if}
              </div>
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <!-- Modal de sélection de type de source -->
  {#if sourceTypeSelectionPending}
    <div class="modal-overlay">
      <div class="modal">
        <h2>Sélectionnez le type de source</h2>
        <div class="source-type-list">
          {#each sourceTypeOptions as sourceType}
            <button
              class="source-type-button"
              on:click={() => handleSourceTypeSelection(sourceType.value)}
            >
              {sourceType.label}
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <!-- Modal de confirmation -->
  {#if confirmPending}
    <div class="modal-overlay">
      <div class="modal">
        <h2>Confirmation</h2>
        <p>{confirmMessage}</p>
        <div class="button-group">
          <button on:click={() => handleConfirmResponse(true)}>Oui</button>
          <button on:click={() => handleConfirmResponse(false)}>Non</button>
        </div>
      </div>
    </div>
  {/if}

  <div class="card">
    <h2>Configuration</h2>

    <div class="form-group">
      <label for="filePath">Fichier vidéo:</label>
      <div style="display: flex; gap: 0.5rem;">
        <input
          id="filePath"
          type="text"
          bind:value={filePath}
          placeholder="Sélectionnez un fichier..."
          readonly
        />
        <button on:click={handleSelectFile}>Parcourir</button>
      </div>
    </div>

    <div class="form-group">
      <label for="outputDir">Dossier de sortie (optionnel):</label>
      <div style="display: flex; gap: 0.5rem;">
        <input
          id="outputDir"
          type="text"
          bind:value={outputDir}
          placeholder="Par défaut: même dossier que le fichier"
        />
        <button on:click={handleSelectDirectory}>Parcourir</button>
      </div>
    </div>

    <div class="form-group">
      <label for="groupName">Nom du groupe:</label>
      <input id="groupName" type="text" bind:value={groupName} />
    </div>

    <div class="form-group checkbox-group">
      <input type="checkbox" id="skipTorrent" bind:checked={skipTorrent} />
      <label for="skipTorrent">Ne pas générer le fichier torrent</label>
    </div>

    <div class="form-group checkbox-group">
      <input type="checkbox" id="noRename" bind:checked={noRename} />
      <label for="noRename">Ne pas renommer le fichier</label>
    </div>

    <div class="button-group">
      <button on:click={handleProcess} disabled={processing || !filePath}>
        {processing ? "⏳ Traitement en cours..." : "🚀 Traiter"}
      </button>
      {#if progressLogs.length > 0}
        <button on:click={clearLogs} disabled={processing}
          >Effacer les logs</button
        >
      {/if}
    </div>
  </div>

  {#if progressLogs.length > 0}
    <div class="card">
      <h2>Progression</h2>
      <div class="progress-log">
        {#each progressLogs as log}
          <div>{log}</div>
        {/each}
      </div>
    </div>
  {/if}

  {#if result}
    <div class="card">
      <h2>Résultat</h2>
      <div class="result-info">
        <p><strong>Film:</strong> {result.movieTitle}</p>
        <p><strong>Nom de release:</strong> {result.releaseName}</p>
        {#if result.nfoPath}
          <p><strong>NFO:</strong> {result.nfoPath}</p>
        {/if}
        {#if result.presentationPath}
          <p><strong>Présentation:</strong> {result.presentationPath}</p>
        {/if}
        {#if result.torrentPath}
          <p><strong>Torrent:</strong> {result.torrentPath}</p>
        {/if}
      </div>
    </div>
  {/if}
</main>

<style>
  main {
    width: 100%;
  }

  h1 {
    color: #1a1a1a;
    font-size: 3em;
    font-weight: 700;
    margin: 0;
  }

  .subtitle {
    color: #666;
    margin-top: 0.5rem;
    margin-bottom: 2rem;
  }

  h2 {
    margin-top: 0;
    margin-bottom: 1rem;
  }

  .result-info p {
    margin: 0.5rem 0;
    word-break: break-all;
  }

  /* Modales */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    max-width: 80%;
    max-height: 80%;
    overflow-y: auto;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  }

  .modal h2 {
    margin-top: 0;
    margin-bottom: 1.5rem;
  }

  /* Liste de films */
  .movie-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .movie-card {
    display: flex;
    gap: 1rem;
    padding: 1rem;
    background: #f5f5f5;
    border: 2px solid #ddd;
    border-radius: 8px;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s;
  }

  .movie-card:hover {
    background: #e8e8e8;
    border-color: #4a9eff;
    transform: translateY(-2px);
  }

  .movie-card img {
    width: 100px;
    height: 150px;
    object-fit: cover;
    border-radius: 4px;
  }

  .movie-info {
    flex: 1;
  }

  .movie-info h3 {
    margin: 0 0 0.5rem 0;
    color: #1a1a1a;
  }

  .original-title {
    font-style: italic;
    color: #666;
    margin: 0.25rem 0;
  }

  .release-date {
    color: #888;
    font-size: 0.9em;
    margin: 0.25rem 0;
  }

  .overview {
    color: #555;
    font-size: 0.9em;
    margin: 0.5rem 0 0 0;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  /* Liste de types de source */
  .source-type-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
  }

  .source-type-button {
    padding: 1.5rem;
    background: #f5f5f5;
    border: 2px solid #ddd;
    border-radius: 8px;
    font-size: 1.1em;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .source-type-button:hover {
    background: #4a9eff;
    border-color: #4a9eff;
    color: white;
    transform: scale(1.05);
  }
</style>
