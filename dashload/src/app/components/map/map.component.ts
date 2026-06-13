import {
  Component,
  OnInit,
  OnDestroy,
  EventEmitter,
  Output,
  Input,
  OnChanges,
  SimpleChanges,
} from '@angular/core';
import * as L from 'leaflet';
import { FlightData, MaritimeData, SatelliteLayer, WeatherData, MapPosition } from '../../models/data.models';
import { SatelliteService } from '../../services/satellite.service';

@Component({
  selector: 'app-map',
  standalone: true,
  template: `<div id="map"></div>`,
  styles: [
    `
      #map {
        width: 100%;
        height: 100%;
      }
    `,
  ],
})
export class MapComponent implements OnInit, OnDestroy, OnChanges {
  @Input() flights: FlightData[] = [];
  @Input() maritimeData: MaritimeData[] = [];
  @Input() satelliteLayers: SatelliteLayer[] = [];
  @Input() weather: WeatherData | null = null;
  @Input() showFlights = true;
  @Input() showMaritime = true;
  @Input() showWeather = true;

  @Output() boundsChanged = new EventEmitter<{
    south: number;
    west: number;
    north: number;
    east: number;
  }>();
  @Output() mapClicked = new EventEmitter<{ lat: number; lng: number }>();
  @Output() positionChanged = new EventEmitter<MapPosition>();

  private map!: L.Map;
  private flightLayerGroup = L.layerGroup();
  private maritimeLayerGroup = L.layerGroup();
  private weatherLayerGroup = L.layerGroup();
  private satelliteTileLayers: Map<string, L.TileLayer> = new Map();

  constructor(private satelliteService: SatelliteService) {}

  ngOnInit(): void {
    this.initMap();
  }

  ngOnDestroy(): void {
    this.map?.remove();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (!this.map) return;

    if (changes['flights'] || changes['showFlights']) {
      this.updateFlightMarkers();
    }
    if (changes['maritimeData'] || changes['showMaritime']) {
      this.updateMaritimeMarkers();
    }
    if (changes['weather'] || changes['showWeather']) {
      this.updateWeatherOverlay();
    }
    if (changes['satelliteLayers']) {
      this.updateSatelliteLayers();
    }
  }

  private initMap(): void {
    this.map = L.map('map', {
      center: [20, 0],
      zoom: 3,
      zoomControl: false,
    });

    L.control.zoom({ position: 'topright' }).addTo(this.map);

    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OSM</a> &copy; <a href="https://carto.com/">CARTO</a>',
      maxZoom: 19,
    }).addTo(this.map);

    this.flightLayerGroup.addTo(this.map);
    this.maritimeLayerGroup.addTo(this.map);
    this.weatherLayerGroup.addTo(this.map);

    this.map.on('moveend', () => this.onBoundsChange());
    this.map.on('click', (e: L.LeafletMouseEvent) => {
      this.mapClicked.emit({ lat: e.latlng.lat, lng: e.latlng.lng });
    });

    // Emit initial bounds
    setTimeout(() => this.onBoundsChange(), 500);
  }

  private onBoundsChange(): void {
    const bounds = this.map.getBounds();
    this.boundsChanged.emit({
      south: bounds.getSouth(),
      west: bounds.getWest(),
      north: bounds.getNorth(),
      east: bounds.getEast(),
    });
    const center = this.map.getCenter();
    this.positionChanged.emit({
      lat: center.lat,
      lng: center.lng,
      zoom: this.map.getZoom(),
    });
  }

  private updateFlightMarkers(): void {
    this.flightLayerGroup.clearLayers();
    if (!this.showFlights) return;

    this.flights.forEach((flight) => {
      const icon = L.divIcon({
        className: 'flight-icon',
        html: `<div class="flight-marker" style="transform: rotate(${flight.heading}deg)">✈</div>`,
        iconSize: [24, 24],
        iconAnchor: [12, 12],
      });

      const marker = L.marker([flight.latitude, flight.longitude], { icon });
      marker.bindPopup(`
        <div class="popup-content">
          <strong>${flight.callsign || flight.icao24}</strong><br>
          Origin: ${flight.originCountry}<br>
          Altitude: ${Math.round(flight.altitude)}m<br>
          Speed: ${Math.round(flight.velocity)} m/s<br>
          ${flight.onGround ? '<span class="on-ground">On Ground</span>' : ''}
        </div>
      `);
      this.flightLayerGroup.addLayer(marker);
    });
  }

  private updateMaritimeMarkers(): void {
    this.maritimeLayerGroup.clearLayers();
    if (!this.showMaritime) return;

    this.maritimeData.forEach((marina) => {
      const icon = L.divIcon({
        className: 'maritime-icon',
        html: `<div class="maritime-marker">⚓</div>`,
        iconSize: [20, 20],
        iconAnchor: [10, 10],
      });

      const marker = L.marker([marina.lat, marina.lng], { icon });
      marker.bindPopup(`
        <div class="popup-content">
          <strong>${marina.name}</strong><br>
          Type: ${marina.type}<br>
        </div>
      `);
      this.maritimeLayerGroup.addLayer(marker);
    });
  }

  private updateWeatherOverlay(): void {
    this.weatherLayerGroup.clearLayers();
    if (!this.showWeather || !this.weather) return;

    const icon = L.divIcon({
      className: 'weather-icon',
      html: `
        <div class="weather-marker">
          <span class="weather-emoji">${this.weather.icon}</span>
          <span class="weather-temp">${Math.round(this.weather.temperature)}°</span>
        </div>
      `,
      iconSize: [60, 40],
      iconAnchor: [30, 20],
    });

    const marker = L.marker([this.weather.lat, this.weather.lng], { icon });
    marker.bindPopup(`
      <div class="popup-content">
        <strong>${this.weather.description}</strong><br>
        Temperature: ${this.weather.temperature}°C<br>
        Humidity: ${this.weather.humidity}%<br>
        Wind: ${this.weather.windSpeed} km/h
      </div>
    `);
    this.weatherLayerGroup.addLayer(marker);
  }

  private updateSatelliteLayers(): void {
    // Remove disabled layers
    this.satelliteTileLayers.forEach((tileLayer, id) => {
      const layer = this.satelliteLayers.find((l) => l.id === id);
      if (!layer || !layer.enabled) {
        this.map.removeLayer(tileLayer);
        this.satelliteTileLayers.delete(id);
      }
    });

    // Add enabled layers
    this.satelliteLayers
      .filter((l) => l.enabled)
      .forEach((layer) => {
        if (!this.satelliteTileLayers.has(layer.id)) {
          const url = this.satelliteService.getTileUrlForDate(layer);
          const tileLayer = L.tileLayer(url, {
            attribution: layer.attribution,
            opacity: layer.opacity,
          });
          tileLayer.addTo(this.map);
          this.satelliteTileLayers.set(layer.id, tileLayer);
        }
      });
  }
}
