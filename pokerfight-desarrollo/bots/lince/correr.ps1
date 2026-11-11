<#
.SYNOPSIS
  Supervisor de Lince: lo lanza y lo vuelve a lanzar si el proceso muere.

.DESCRIPTION
  El bot ya reconecta solo si se corta la red. Este script cubre el otro caso:
  que el proceso entero se caiga. Lo relanza con el mismo token, y el bot
  retoma del disco en que ronda iba.

  Solo deja correr una instancia: dos conexiones con el mismo token se pisan
  (la segunda le quita la silla a la primera).

  Se detiene con Ctrl+C.

  (Sin tildes a proposito: Windows PowerShell 5.1 lee los .ps1 sin BOM como ANSI.)

.EXAMPLE
  .\bots\lince\correr.ps1 -Addr 32.201.29.142:9000 -Token pfb-xxxx -Nuevo

.EXAMPLE
  $env:POKERFIGHT_TOKEN = 'pfb-xxxx'; .\bots\lince\correr.ps1 -Addr host:9000
#>
param(
    # host:puerto del servidor.
    [Parameter(Mandatory = $true)][string]$Addr,
    # Token del equipo. Si no se pasa, se toma de la variable POKERFIGHT_TOKEN.
    [string]$Token = $env:POKERFIGHT_TOKEN,
    # Carpeta del estado y los logs. Por defecto, bots\lince\datos.
    [string]$Datos,
    # auto en el torneo; clasificatoria en el servidor de pruebas.
    [ValidateSet('auto', 'clasificatoria', 'final')][string]$Fase = 'auto',
    # Aparta el estado guardado y arranca de cero. Usarlo UNA vez, al empezar un
    # torneo. Nunca al relanzar a mitad de un torneo: se pierde la cuenta de rondas.
    [switch]$Nuevo,
    # Ejecutable del bot. Por defecto, bin\lince.exe del repositorio.
    [string]$Exe,
    # No imprime cada decision en la terminal (al archivo de log van igual).
    [switch]$Silencioso
)

$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
if (-not $Exe) { $Exe = Join-Path $repo 'bin\lince.exe' }
if (-not $Datos) { $Datos = Join-Path $PSScriptRoot 'datos' }
if (-not $Token) { throw 'Falta el token: pasar -Token o definir la variable POKERFIGHT_TOKEN.' }

New-Item -ItemType Directory -Force $Datos | Out-Null
$registro = Join-Path $Datos 'supervisor.log'

function Anotar([string]$texto) {
    $linea = '{0} supervisor: {1}' -f (Get-Date -Format 'yyyy/MM/dd HH:mm:ss'), $texto
    Write-Host $linea
    Add-Content -Path $registro -Value $linea -Encoding utf8
}

# Una sola instancia en toda la maquina.
$creado = $false
$candado = New-Object System.Threading.Mutex($true, 'Global\PokerFightLince', [ref]$creado)
if (-not $creado) {
    Write-Host 'Ya hay un supervisor de Lince corriendo en esta maquina. No se lanza otro.'
    exit 3
}

try {
    if ($Nuevo) {
        $estado = Join-Path $Datos 'estado.json'
        if (Test-Path $estado) {
            $destino = Join-Path $Datos ('estado.manual-{0}.json' -f (Get-Date -Format 'yyyyMMdd-HHmmss'))
            Move-Item $estado $destino
            Anotar "estado anterior apartado en $destino"
        }
    }

    $argumentos = @('-addr', $Addr, '-token', $Token, '-datos', $Datos, '-fase', $Fase)
    if ($Silencioso) { $argumentos += '-silencioso' }
    $lanzamiento = 0
    while ($true) {
        $lanzamiento++
        $inicio = Get-Date
        $codigo = $null
        Anotar "lanzamiento $lanzamiento de $Exe"
        try {
            & $Exe @argumentos
            $codigo = $LASTEXITCODE
        }
        catch {
            # Tipico: Smart App Control bloqueo el .exe. Se intenta con go run.
            Anotar "no se pudo ejecutar el .exe ($($_.Exception.Message)); se intenta con go run"
            Push-Location $repo
            try {
                & go run ./bots/lince @argumentos
                $codigo = $LASTEXITCODE
            }
            catch { Anotar "go run tambien fallo: $($_.Exception.Message)" }
            finally { Pop-Location }
        }
        $duro = [int]((Get-Date) - $inicio).TotalSeconds
        # Si murio enseguida, algo esta mal de entrada (un flag, el genoma): se
        # espera mas para no llenar la pantalla ni martillar al servidor.
        $espera = 1
        if ($duro -lt 5) { $espera = 5 }
        Anotar "el bot termino (codigo $codigo) tras $duro s; se relanza en $espera s"
        Start-Sleep -Seconds $espera
    }
}
finally {
    $candado.ReleaseMutex()
    $candado.Dispose()
}
