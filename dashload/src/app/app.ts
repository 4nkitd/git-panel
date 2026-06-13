import { Component, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subject, debounceTime, takeUntil } from 'rxjs';

import { HeaderComponent } from './components/header/header.component';
import { SidebarComponent } from './components/sidebar/sidebar.component';
import { MapComponent } from './components/map/map.component';
import { DataPanelComponent } from './components/data-panel/data-panel.component';

import { FlightService } from './services/flight.service';
import { MaritimeService } from './services/maritime.service';
import { WeatherService } from './services/weather.service';
import { NewsService } from './services/news.service';
import { SatelliteService } from './services/satellite.service';

import {
  FlightData,
  MaritimeData,
  WeatherData,
  NewsArticle,
  SatelliteLayer,
  LayerToggle,
  DataCategory,
  MapPosition,
} from './models/data.models';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    HeaderComponent,
    SidebarComponent,
    MapComponent,
    DataPanelComponent,
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss',
})
export class App implements OnDestroy {
  flights: FlightData[] = [];
  maritimeData: MaritimeData[] = [];
  weather: WeatherData | null = null;
  news: NewsArticle[] = [];
  satelliteLayers: SatelliteLayer[] = [];

  loading = false;
  currentLat: number | null = null;
  currentLng: number | null = null;
  private currentCategory = '';

  layers: LayerToggle[] = [
    { category: 'flights', label: 'Flights', icon: '✈️', enabled: true },
    { category: 'maritime', label: 'Maritime', icon: '⚓', enabled: true },
    { category: 'weather', label: 'Weather', icon: '🌤️', enabled: true },
    { category: 'news', label: 'News', icon: '📰', enabled: true },
    { category: 'satellite', label: 'Satellite', icon: '🛰️', enabled: false },
  ];

  private boundsSubject = new Subject<{
    south: number;
    west: number;
    north: number;
    east: number;
  }>();
  private destroy$ = new Subject<void>();

  constructor(
    private flightService: FlightService,
    private maritimeService: MaritimeService,
    private weatherService: WeatherService,
    private newsService: NewsService,
    private satelliteService: SatelliteService
  ) {
    this.satelliteLayers = this.satelliteService.getLayers();

    this.boundsSubject
      .pipe(debounceTime(800), takeUntil(this.destroy$))
      .subscribe((bounds) => this.loadDataForBounds(bounds));
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  isLayerEnabled(category: DataCategory): boolean {
    return this.layers.find((l) => l.category === category)?.enabled ?? false;
  }

  onBoundsChanged(bounds: {
    south: number;
    west: number;
    north: number;
    east: number;
  }): void {
    this.boundsSubject.next(bounds);
  }

  onMapClicked(pos: { lat: number; lng: number }): void {
    this.currentLat = pos.lat;
    this.currentLng = pos.lng;
    this.loadPointData(pos.lat, pos.lng);
  }

  onPositionChanged(pos: MapPosition): void {
    this.currentLat = pos.lat;
    this.currentLng = pos.lng;
  }

  onLayerToggled(_category: DataCategory): void {
    // Layers are toggled in-place by sidebar; trigger a refresh
    this.layers = [...this.layers];
  }

  onSatelliteToggled(): void {
    this.satelliteLayers = [...this.satelliteLayers];
  }

  onCategoryChanged(category: string): void {
    this.currentCategory = category;
    if (this.currentLat !== null && this.currentLng !== null) {
      this.loadNews(this.currentLat, this.currentLng);
    }
  }

  onLocationSelected(loc: {
    name: string;
    lat: number;
    lng: number;
    zoom: number;
  }): void {
    this.currentLat = loc.lat;
    this.currentLng = loc.lng;
    this.loadPointData(loc.lat, loc.lng);
  }

  onSearch(query: string): void {
    // Try to parse as coordinates (lat, lng)
    const coordMatch = query.match(
      /^(-?\d+\.?\d*)\s*,\s*(-?\d+\.?\d*)$/
    );
    if (coordMatch) {
      const lat = parseFloat(coordMatch[1]);
      const lng = parseFloat(coordMatch[2]);
      this.currentLat = lat;
      this.currentLng = lng;
      this.loadPointData(lat, lng);
      return;
    }

    // Otherwise search news by keyword
    this.loading = true;
    this.newsService
      .getNewsByKeyword(query)
      .pipe(takeUntil(this.destroy$))
      .subscribe((articles) => {
        this.news = articles;
        this.loading = false;
        this.updateLayerCount('news', articles.length);
      });
  }

  private loadDataForBounds(bounds: {
    south: number;
    west: number;
    north: number;
    east: number;
  }): void {
    this.loading = true;

    if (this.isLayerEnabled('flights')) {
      this.flightService
        .getFlightsInBounds(
          bounds.south,
          bounds.west,
          bounds.north,
          bounds.east
        )
        .pipe(takeUntil(this.destroy$))
        .subscribe((data) => {
          this.flights = data;
          this.updateLayerCount('flights', data.length);
          this.checkLoadingDone();
        });
    }

    if (this.isLayerEnabled('maritime')) {
      this.maritimeService
        .getMarinasInBounds(
          bounds.south,
          bounds.west,
          bounds.north,
          bounds.east
        )
        .pipe(takeUntil(this.destroy$))
        .subscribe((data) => {
          this.maritimeData = data;
          this.updateLayerCount('maritime', data.length);
          this.checkLoadingDone();
        });
    }

    this.checkLoadingDone();
  }

  private loadPointData(lat: number, lng: number): void {
    this.loading = true;

    if (this.isLayerEnabled('weather')) {
      this.weatherService
        .getWeather(lat, lng)
        .pipe(takeUntil(this.destroy$))
        .subscribe((data) => {
          this.weather = data;
          this.checkLoadingDone();
        });
    }

    if (this.isLayerEnabled('news')) {
      this.loadNews(lat, lng);
    }

    this.checkLoadingDone();
  }

  private loadNews(lat: number, lng: number): void {
    this.newsService
      .getNewsByLocation(lat, lng, this.currentCategory || undefined)
      .pipe(takeUntil(this.destroy$))
      .subscribe((articles) => {
        this.news = articles;
        this.updateLayerCount('news', articles.length);
        this.checkLoadingDone();
      });
  }

  private updateLayerCount(category: DataCategory, count: number): void {
    const layer = this.layers.find((l) => l.category === category);
    if (layer) {
      layer.count = count;
      this.layers = [...this.layers];
    }
  }

  private checkLoadingDone(): void {
    this.loading = false;
  }
}
