# Epic Fain

Plataforma IoT robusta para la recopilación y gestión de telemetría en tiempo real mediante comunicación por bus CAN. Construida en Go y diseñada para escalabilidad, confiabilidad y monitoreo de cumplimiento en equipos industriales.

## Descripción General

Epic Fain es una aplicación backend lista para microservicios que recopila, procesa y gestiona datos de telemetría de dispositivos de la serie Epic. Maneja análisis de frames CAN, monitoreo en tiempo real, registro de auditoría y generación de alertas con una arquitectura hexagonal limpia.

### Características Principales

- **Integración CAN Bus**: Recepción y análisis en tiempo real de frames CAN desde dispositivos remotos
- **Gestión de Telemetría**: Almacenar y recuperar mediciones de dispositivos (voltaje, corriente, potencia, datos de batería de litio)
- **Registro de Auditoría**: Pista de auditoría completa de todas las operaciones del sistema para cumplimiento normativo
- **Sistema de Alertas**: Generación automática de alertas para anomalías y violaciones de umbrales
- **Control de Dispositivos**: Comandar dispositivos a través de la plataforma (Hito 3)
- **API REST**: API HTTP completa para interacción con el sistema
- **Multi-protocolo**: Comunicación dual vía HTTP y TCP
- **Persistencia en Base de Datos**: PostgreSQL para almacenamiento confiable de datos y políticas de retención

## Stack Tecnológico

- **Lenguaje**: Go 1.22
- **Base de Datos**: PostgreSQL 16
- **Contenedorización**: Docker & Docker Compose
- **Arquitectura**: Hexagonal (Puertos & Adaptadores)
- **Dependencias Clave**:
  - `github.com/lib/pq` - Driver PostgreSQL
  - `github.com/google/uuid` - Generación de UUID

## Arquitectura

El proyecto sigue el patrón de **Arquitectura Hexagonal** (Puertos & Adaptadores):

```
epic-fain/
├── cmd/
│   └── server/          # Punto de entrada de la aplicación
├── internal/
│   ├── domain/          # Lógica de negocio central (independiente de infraestructura)
│   │   ├── model/       # Entidades del dominio (CANFrame, Telemetry, Alert, Audit, etc.)
│   │   ├── port/        # Definición de interfaces (Repository, Event, etc.)
│   │   └── service/     # Servicios de dominio (decodificador CAN, reglas de negocio)
│   ├── application/     # Servicios de aplicación orquestando lógica de dominio
│   │   ├── device_control_service.go
│   │   └── telemetry_service.go
│   └── infrastructure/  # Adaptadores externos
│       ├── adapter/
│       │   ├── inbound/     # Servidores HTTP y TCP
│       │   └── outbound/    # Repositorios de base de datos, servicios externos
│       ├── config/          # Gestión de configuración
│       └── migration/       # Esquemas de base de datos
```

### Modelo de Dominio

Entidades clave gestionadas por la plataforma:

- **CANFrame**: Mensajes CAN brutos con MessageID, DLC y datos
- **Telemetry**: Mediciones de dispositivos (estado, voltaje, corriente, potencia, datos de litio)
- **Command**: Directivas de control enviadas a dispositivos (control VVVF, config de mediciones)
- **Alert**: Alertas generadas para anomalías y violaciones de umbrales
- **AuditLog**: Registro completo de operaciones de la plataforma
- **Installation**: Configuración y metadatos de dispositivos/instalaciones

### Protocolo de Mensajes CAN

#### Entrantes (Dispositivos → Plataforma)
- `0xFF00` - Estado (7 bytes)
- `0xFF01` - Mediciones (8 bytes)
- `0xFF02` - Corrientes (4 bytes)
- `0xFF03` - Datos de litio (8 bytes, solo dispositivos habilitados para litio)
- `0xFF0E` - Información del dispositivo (8 bytes)

#### Salientes (Plataforma → Dispositivos)
- `0xEF00` - Control VVVF/Reset (1 byte)
- `0xEF01` - Configuración de mediciones (3 bytes)
- `0xEF0E` - Solicitud de información del dispositivo (1 byte)

## Inicio Rápido

### Requisitos Previos

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16+ (o usar Docker)

### Instalación

1. **Clonar el repositorio**
   ```bash
   git clone https://github.com/liftelimasd/epic-fain.git
   cd epic-fain
   ```

2. **Instalar dependencias**
   ```bash
   go mod download
   ```

3. **Iniciar con Docker Compose** (recomendado)
   ```bash
   docker-compose up
   ```
   Esto inicia tanto PostgreSQL como la aplicación Epic Fain.

4. **O ejecutar localmente**
   ```bash
   # Asegurate de que PostgreSQL está ejecutándose
   # Establecer variables de entorno
   export HTTP_ADDR=":8080"
   export TCP_ADDR=":9090"
   export DB_HOST="localhost"
   export DB_PORT="5432"
   export DB_USER="epicfain"
   export DB_PASSWORD="epicfain"
   export DB_NAME="epicfain"
   export DB_SSLMODE="disable"
   
   go run ./cmd/server
   ```

### Variables de Entorno

| Variable | Predeterminado | Descripción |
|----------|---|-------------|
| `HTTP_ADDR` | `:8080` | Dirección de escucha del servidor HTTP |
| `TCP_ADDR` | `:9090` | Dirección de escucha del servidor TCP (receptor de datos CAN) |
| `DB_HOST` | `localhost` | Nombre de host de PostgreSQL |
| `DB_PORT` | `5432` | Puerto de PostgreSQL |
| `DB_USER` | `epicfain` | Usuario de PostgreSQL |
| `DB_PASSWORD` | `epicfain` | Contraseña de PostgreSQL |
| `DB_NAME` | `epicfain` | Nombre de la base de datos PostgreSQL |
| `DB_SSLMODE` | `disable` | Modo SSL de PostgreSQL |
| `API_KEYS` | (opcional) | Claves API separadas por comas para autenticación |

## Endpoints de API

Todos los endpoints HTTP requieren autenticación por clave API (encabezado: `X-API-Key`).

### Telemetría
- `GET /api/v1/telemetry` - Recuperar datos de telemetría
- `POST /api/v1/telemetry` - Crear nuevo registro de telemetría

### Auditoría
- `GET /api/v1/audit` - Recuperar logs de auditoría

### Dispositivos
- `GET /api/v1/devices` - Listar dispositivos (Hito 2+)
- `POST /api/v1/devices/{id}/command` - Enviar comando a dispositivo (Hito 3)

### Estado e Información
- `GET /health` - Endpoint de verificación de salud
- `GET /api/v1/info` - Información de la plataforma

## Protocolo TCP

La plataforma escucha en el puerto TCP para datos de frames CAN. Los dispositivos deben enviar frames CAN brutos en el siguiente formato binario:

```
[MessageID: 2 bytes][DLC: 1 byte][Data: hasta 8 bytes]
```

## Esquema de Base de Datos

Las migraciones automáticas se ejecutan al iniciar:

1. **001_initial_schema.sql** - Tablas para telemetría, auditoría, alertas, instalaciones, comandos
2. **002_retention_policy.sql** - Reglas y políticas de retención de datos

## Testing

Ejecutar la suite de pruebas:

```bash
go test ./...
```

Ejecutar con cobertura:

```bash
go test -cover ./...
```

Archivos de prueba clave:
- `internal/domain/service/decoder_test.go` - Pruebas del decodificador CAN

## Desarrollo

### Estilo de Código

Sigue convenciones e idiomas de Go:
- Usar pruebas table-driven
- Implementar manejo adecuado de errores
- Mantener paquetes enfocados y modulares
- Respetar los límites de arquitectura

### Skills Recomendadas

El proyecto incluye skills Go automatizadas:
- **Go Development Patterns** - Patrones idiomáticos y mejores prácticas
- **Go Testing Patterns** - Metodología TDD con pruebas table-driven

Accesibles en `.claude/skills/` o `.agents/skills/`.

### Fases del Proyecto

- **Hito 1**: Recopilación de telemetría central y API ✅
- **Hito 2**: Gestión de dispositivos, seguimiento de instalaciones
- **Hito 3**: Sistema de alertas, control de dispositivos, integración MQTT

## Despliegue

### Docker

1. **Construir la imagen**
   ```bash
   docker build -t liftel/epic-fain:latest .
   ```

2. **Ejecutar con compose**
   ```bash
   docker-compose up -d
   ```

3. **Ver logs**
   ```bash
   docker-compose logs -f epic-fain
   ```

4. **Detener servicios**
   ```bash
   docker-compose down
   ```

### Verificación de Salud

La aplicación expone un endpoint de salud:
```bash
curl http://localhost:8080/health
```

## Migraciones de Base de Datos

Las migraciones se aplican automáticamente al iniciar desde archivos SQL en `internal/infrastructure/migration/`.

Para ejecutar migraciones manualmente:
```bash
# Asegurate de que las herramientas CLI de PostgreSQL estén instaladas
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < internal/infrastructure/migration/001_initial_schema.sql
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < internal/infrastructure/migration/002_retention_policy.sql
```

## Solución de Problemas

### Errores de conexión a base de datos
- Verifica que PostgreSQL esté ejecutándose: `docker-compose ps`
- Comprueba que las credenciales de la base de datos coincidan con las variables de entorno
- Asegurate de que la base de datos esté sana: `docker-compose logs db`

### El servidor TCP no recibe datos
- Verifica que TCP_ADDR esté configurado correctamente
- Comprueba las reglas del firewall permitan conexiones al puerto TCP
- Asegurate de que los dispositivos remotos envíen al host/puerto correcto

### Fallos de autenticación de API
- Verifica que la variable de entorno `API_KEYS` esté establecida
- Asegurate de que el encabezado `X-API-Key` esté incluido en las solicitudes
- Comprueba el formato de la clave API (separadas por comas si hay múltiples)

## Contribuir

1. Fork el repositorio
2. Crea una rama de características: `git checkout -b feature/tu-feature`
3. Commit tus cambios con mensajes claros
4. Haz push a tu fork
5. Envía una pull request

## Licencia

[Especifica tu licencia aquí - ej: MIT, Apache 2.0, etc.]

## Soporte

Para issues, preguntas o feedback:
- Abre un issue en GitHub: https://github.com/liftelimasd/epic-fain/issues
- Contacta al equipo de Liftel

## Información del Proyecto

- **Repositorio**: https://github.com/liftelimasd/epic-fain
- **Organización**: [Liftel](https://github.com/liftelimasd)
- **Lenguaje**: Go 1.22
- **Última Actualización**: 2026
