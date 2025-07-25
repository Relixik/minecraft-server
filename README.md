
# 🎮 GoLangMC - Serveur Minecraft en Go

<p align="center">
  <a href="https://github.com/GoLangMc/minecraft-server">
    <img src="https://avatars3.githubusercontent.com/u/61735329" alt="logo" width="100" height="100">
  </a>
</p>

<p align="center">
  <strong>Une implémentation moderne d'un serveur Minecraft écrite en Go</strong>
  <br />
  <em>Performance, concurrence et simplicité</em>
</p>

<p align="center">
  <a href="#-fonctionnalités-implémentées">Fonctionnalités</a> •
  <a href="#-architecture">Architecture</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-optimisations-possibles">Optimisations</a> •
  <a href="#-feuille-de-route">Roadmap</a>
</p>

---

## 📋 Vue d'ensemble

Ce projet est une implémentation complète d'un serveur Minecraft utilisant le protocole officiel de Minecraft Java Edition. Il est conçu pour être performant, modulaire et facilement extensible grâce aux capacités de concurrence native de Go.

### 🎯 Objectifs du projet
- **Performance** : Exploitation maximale du multi-threading Go
- **Compatibilité** : Support du protocole Minecraft officiel
- **Modularité** : Architecture claire et extensible
- **Simplicité** : Code lisible et maintenable

## 🚀 Fonctionnalités implémentées

### ✅ Système de réseau et protocole
- **États de connexion** : Handshake, Status, Login, Play
- **Gestion des paquets** : Système complet d'encodage/décodage
- **Chiffrement** : Implémentation CFB8 avec RSA pour l'authentification
- **Compression** : Support de la compression des paquets
- **Keep-alive** : Maintien des connexions actives

### ✅ Authentification et sécurité
- **Authentification Mojang** : Vérification des comptes via l'API officielle
- **Chiffrement AES** : Communication sécurisée client-serveur
- **Vérification des tokens** : Validation des clés d'authentification
- **Profils joueurs** : Gestion des UUID et propriétés (skins, capes)

### ✅ Gestion des joueurs
- **Connexion/Déconnexion** : Événements complets de session
- **Métadonnées** : Gestion des informations joueur (nom, UUID, skin)
- **Modes de jeu** : Support des différents GameModes
- **Capacités** : Gestion du vol, invulnérabilité, etc.
- **Position et rotation** : Tracking complet des mouvements

### ✅ Système de monde
- **Chunks** : Génération et gestion des chunks 16x16x256
- **Slices** : Découpage vertical en sections de 16 blocs
- **Blocs** : Système de placement et récupération
- **Génération SuperFlat** : Générateur de monde plat basique
- **Height Maps** : Cartes de hauteur pour l'optimisation

### ✅ Chat et commandes
- **Système de chat** : Messages formatés avec couleurs
- **Commandes console** : `stop`, `send`, `vers`
- **Broadcasting** : Diffusion de messages à tous les joueurs
- **Formatage** : Support des codes couleur Minecraft

### ✅ Système de tâches
- **Scheduler** : Exécution de tâches périodiques et différées
- **Multi-threading** : Gestion asynchrone des tâches
- **Keep-alive automatique** : Maintenance des connexions

### ✅ Plugin et extensibilité
- **Messages de plugin** : Canaux de communication personnalisés
- **Brand detection** : Détection du client utilisé
- **Système d'événements** : Architecture event-driven

## 🏗️ Architecture

### Structure modulaire
```
apis/          → Interfaces et contrats publics
├── server/    → Interface principale du serveur
├── ents/      → Entités (joueurs, mobs)
├── game/      → Éléments de jeu (blocs, chunks, modes)
├── cmds/      → Système de commandes
├── logs/      → Système de logging
└── task/      → Gestionnaire de tâches

impl/          → Implémentations concrètes
├── server/    → Logique principale du serveur
├── conn/      → Gestion des connexions réseau
├── prot/      → Implémentation du protocole
├── game/      → Logique de jeu et monde
└── auth/      → Authentification et sécurité
```

### Flux de données
```
Client ←→ Network ←→ Packets ←→ Game Logic ←→ World State
                ↓
            Authentication ←→ Mojang API
                ↓
            Player Management ←→ Events
```

## 🔧 Installation

### Prérequis
- Go 1.13 ou supérieur
- Connexion Internet (pour l'authentification Mojang)

### Compilation et lancement
```bash
# Cloner le repository
git clone https://github.com/GoLangMc/minecraft-server.git
cd minecraft-server

# Installer les dépendances
go mod download

# Compiler et lancer
go run main.go

# Ou avec des paramètres personnalisés
go run main.go -host 0.0.0.0 -port 25565
```

### Configuration
Le serveur utilise actuellement une configuration par défaut :
- **Host** : `0.0.0.0`
- **Port** : `25565`
- **Mode** : Creative
- **Difficulté** : Peaceful
- **Monde** : SuperFlat

## ⚡ Optimisations possibles

### 🧵 Multi-threading et concurrence

#### Améliorations actuelles à implémenter :
1. **Pool de workers pour les chunks**
   ```go
   // Paralléliser la génération de chunks
   type ChunkWorkerPool struct {
       workers    int
       chunkQueue chan ChunkRequest
       resultChan chan ChunkResult
   }
   ```

2. **Goroutines par joueur**
   ```go
   // Une goroutine dédiée par connexion joueur
   go handlePlayerConnection(player, conn)
   ```

3. **Lock-free data structures**
   - Utiliser `sync.Map` pour les associations joueur-connexion
   - Channels pour la communication inter-goroutines
   - Atomic operations pour les compteurs

#### Propositions d'architecture multi-thread :
```go
// Serveur principal
type Server struct {
    networkPool    *WorkerPool  // Pool pour réseau
    chunkPool      *WorkerPool  // Pool pour chunks  
    playerPool     *WorkerPool  // Pool pour joueurs
    eventBus       *EventBus    // Bus d'événements async
}
```

### 💾 Optimisations mémoire

#### 1. Object pooling
```go
// Pool de buffers pour éviter les allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}
```

#### 2. Gestion intelligente des chunks
```go
// Déchargement automatique des chunks non utilisés
type ChunkManager struct {
    loadedChunks  map[ChunkPos]*Chunk
    lastAccess    map[ChunkPos]time.Time
    unloadTimer   *time.Timer
}
```

#### 3. Compression des données
- Compression des chunks stockés en mémoire
- Delta compression pour les mises à jour de position
- Sérialisation binaire optimisée

### 🚀 Optimisations de performance

#### 1. Mise en cache intelligente
```go
type CacheLayer struct {
    chunkCache    *LRUCache
    playerCache   *LRUCache
    packetCache   *LRUCache
}
```

#### 2. Batch processing
```go
// Traitement par lots des paquets
type PacketBatcher struct {
    packets []Packet
    timer   *time.Timer
    flush   func([]Packet)
}
```

#### 3. Profiling intégré
```go
// Métriques de performance en temps réel
type ServerMetrics struct {
    PacketsPerSecond  int64
    PlayersOnline     int32
    ChunksLoaded      int32
    MemoryUsage       int64
}
```

## 📈 Fonctionnalités à implémenter

### 🎯 Priorité haute

#### 1. Système d'inventaire complet
```go
type Inventory struct {
    Slots     []*ItemStack
    Size      int
    HotbarIdx int
}

type ItemStack struct {
    Material Material
    Count    int
    NBT      *NBTCompound
}
```

#### 2. Entités et mobs
```go
type Entity interface {
    Tick()
    GetPosition() Position
    SetPosition(Position)
    GetVelocity() Vector3
    OnCollision(Entity)
}

type Mob interface {
    Entity
    GetHealth() float64
    SetHealth(float64)
    GetAI() AI
}
```

#### 3. Redstone et mécaniques
```go
type RedstoneComponent interface {
    GetPowerLevel() int
    SetPowerLevel(int)
    UpdatePower()
}

type BlockState struct {
    Material   Material
    Properties map[string]interface{}
    Redstone   RedstoneComponent
}
```

#### 4. Génération de monde avancée
```go
type WorldGenerator interface {
    GenerateChunk(x, z int) *Chunk
    GetBiome(x, z int) Biome
    GenerateOres(chunk *Chunk)
    GenerateStructures(chunk *Chunk)
}

// Générateurs spécialisés
type NoiseGenerator struct {
    Seed   int64
    Octaves []NoiseOctave
}
```

### 🎯 Priorité moyenne

#### 1. Système de plugins robuste
```go
type Plugin interface {
    OnEnable() error
    OnDisable() error
    OnPlayerJoin(event PlayerJoinEvent)
    OnBlockPlace(event BlockPlaceEvent)
}

type PluginManager struct {
    plugins    map[string]Plugin
    eventBus   *EventBus
    scheduler  *TaskScheduler
}
```

#### 2. Base de données et persistance
```go
type PlayerData struct {
    UUID      string
    Inventory *Inventory
    Position  Position
    Stats     map[string]int64
}

type WorldStorage interface {
    SaveChunk(*Chunk) error
    LoadChunk(x, z int) (*Chunk, error)
    SavePlayerData(*PlayerData) error
    LoadPlayerData(uuid string) (*PlayerData, error)
}
```

#### 3. Interface d'administration web
```go
type WebAdmin struct {
    server    *http.Server
    wsHandler *websocket.Handler
    auth      AuthProvider
}

// API REST pour management
type AdminAPI struct {
    Players    PlayersEndpoint
    Server     ServerEndpoint
    Worlds     WorldsEndpoint
    Plugins    PluginsEndpoint
}
```

#### 4. Système de permissions
```go
type Permission struct {
    Node        string
    Default     bool
    Description string
}

type PermissionProvider interface {
    HasPermission(player Player, node string) bool
    GrantPermission(player Player, node string)
    RevokePermission(player Player, node string)
}
```

### 🎯 Priorité basse

#### 1. Support multi-monde
```go
type MultiWorldManager struct {
    worlds      map[string]*World
    defaultWorld string
    generators  map[string]WorldGenerator
}
```

#### 2. Économie intégrée
```go
type Economy interface {
    GetBalance(player Player) float64
    Deposit(player Player, amount float64) bool
    Withdraw(player Player, amount float64) bool
    Transfer(from, to Player, amount float64) bool
}
```

## 🔄 Mises à jour nécessaires

### 📦 Dépendances
- **Go version** : Migrer vers Go 1.21+ pour les génériques
- **Bibliothèques** : Mise à jour vers les versions récentes
- **UUID library** : Remplacer par `google/uuid`

### 🔧 Refactoring suggéré

#### 1. Utilisation des génériques Go
```go
// Avant
type PlayerMap map[uuid.UUID]ents.Player

// Après (avec génériques)
type SafeMap[K comparable, V any] struct {
    mu sync.RWMutex
    m  map[K]V
}
```

#### 2. Context pour la gestion des timeouts
```go
func (s *Server) HandleConnection(ctx context.Context, conn net.Conn) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Logique de connexion avec timeout
}
```

#### 3. Structured logging
```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("Player joined", 
    "player", player.Name(),
    "uuid", player.UUID(),
    "ip", conn.RemoteAddr(),
)
```

## 🛠️ Développement et contribution

### Tests
```bash
# Lancer les tests unitaires
go test ./...

# Tests avec couverture
go test -cover ./...

# Benchmarks
go test -bench=. ./...
```

### Profiling
```bash
# Profile CPU
go test -cpuprofile=cpu.prof -bench=.

# Profile mémoire  
go test -memprofile=mem.prof -bench=.

# Analyse avec pprof
go tool pprof cpu.prof
```

### Standards de code
- Suivre `gofmt` et `goimports`
- Utiliser `golangci-lint` pour la qualité
- Documenter toutes les fonctions publiques
- Tests unitaires obligatoires pour les nouvelles fonctionnalités

## 📊 Métriques de performance

### Objectifs de performance
- **Connexions simultanées** : 500+ joueurs
- **TPS** : 20 ticks constants
- **Latence** : < 50ms pour les actions locales
- **Mémoire** : < 2GB pour 100 joueurs
- **CPU** : < 80% sur 4 cores

### Monitoring
```go
type ServerStats struct {
    PlayersOnline    int32
    TPS             float64
    MemoryUsage     int64
    PacketsSent     int64
    PacketsReceived int64
    ChunksLoaded    int32
}
```

## 🚧 Limitations actuelles

1. **Pas de persistance** : Données perdues au redémarrage
2. **Génération basique** : Seulement SuperFlat
3. **Pas d'entités** : Aucun mob ou animal
4. **Inventaire simplifié** : Système incomplet
5. **Un seul monde** : Pas de support multi-monde
6. **Pas de plugins** : Système d'extension limité

## 📝 Notes techniques

### Protocol version
- **Actuellement supporté** : Version non spécifiée (semble être 1.14+)
- **À préciser** : Quelle version exacte de Minecraft

### Compatibilité client
- Clients Minecraft Java Edition officiels
- Possibilité d'étendre pour d'autres clients

## 🤝 Contribution

Les contributions sont les bienvenues ! Voir les issues GitHub pour les tâches prioritaires.

### Guide de contribution
1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/AmazingFeature`)
3. Commit les changements (`git commit -m 'Add some AmazingFeature'`)
4. Push la branche (`git push origin feature/AmazingFeature`)
5. Ouvrir une Pull Request

---

<p align="center">
  <strong>Développé avec ❤️ par la communauté GoLangMC</strong>
</p>

<p align="center">
  <a href="https://github.com/GoLangMc/minecraft-server/issues">🐛 Report Bug</a> •
  <a href="https://github.com/GoLangMc/minecraft-server/issues">💡 Request Feature</a> •
  <a href="https://github.com/GoLangMc/minecraft-server/pulls">🔄 Pull Request</a>
</p>