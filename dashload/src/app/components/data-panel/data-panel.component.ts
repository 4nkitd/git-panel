import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NewsArticle, WeatherData, FlightData, MaritimeData } from '../../models/data.models';

@Component({
  selector: 'app-data-panel',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="data-panel" [class.expanded]="expanded">
      <div class="panel-handle" (click)="expanded = !expanded">
        <span class="handle-bar"></span>
        <span class="panel-title">{{ expanded ? 'Data Panel' : 'Show Data' }}</span>
      </div>

      <div class="panel-content" *ngIf="expanded">
        <!-- Weather Card -->
        <div class="data-card weather-card" *ngIf="weather">
          <div class="card-header">
            <span class="card-icon">{{ weather.icon }}</span>
            <h4>Weather</h4>
          </div>
          <div class="weather-info">
            <div class="temp">{{ weather.temperature }}°C</div>
            <div class="weather-details">
              <span>{{ weather.description }}</span>
              <span>Humidity: {{ weather.humidity }}%</span>
              <span>Wind: {{ weather.windSpeed }} km/h</span>
            </div>
          </div>
        </div>

        <!-- Flight Stats -->
        <div class="data-card" *ngIf="flights.length > 0">
          <div class="card-header">
            <span class="card-icon">✈️</span>
            <h4>Flights in View</h4>
            <span class="badge">{{ flights.length }}</span>
          </div>
          <div class="flight-list">
            <div class="flight-item" *ngFor="let f of flights | slice:0:10">
              <span class="callsign">{{ f.callsign || f.icao24 }}</span>
              <span class="origin">{{ f.originCountry }}</span>
              <span class="alt">{{ (f.altitude | number:'1.0-0') }}m</span>
            </div>
            <div class="more-indicator" *ngIf="flights.length > 10">
              +{{ flights.length - 10 }} more
            </div>
          </div>
        </div>

        <!-- Maritime -->
        <div class="data-card" *ngIf="maritimeData.length > 0">
          <div class="card-header">
            <span class="card-icon">⚓</span>
            <h4>Marinas & Ports</h4>
            <span class="badge">{{ maritimeData.length }}</span>
          </div>
          <div class="marina-list">
            <div class="marina-item" *ngFor="let m of maritimeData | slice:0:8">
              <span class="marina-name">{{ m.name }}</span>
              <span class="marina-type">{{ m.type }}</span>
            </div>
          </div>
        </div>

        <!-- News -->
        <div class="data-card news-card" *ngIf="news.length > 0">
          <div class="card-header">
            <span class="card-icon">📰</span>
            <h4>News</h4>
            <span class="badge">{{ news.length }}</span>
          </div>
          <div class="news-list">
            <a
              class="news-item"
              *ngFor="let article of news | slice:0:8"
              [href]="article.url"
              target="_blank"
              rel="noopener"
            >
              <img
                *ngIf="article.imageUrl"
                [src]="article.imageUrl"
                class="news-thumb"
                loading="lazy"
                (error)="article.imageUrl = undefined"
              />
              <div class="news-text">
                <span class="news-title">{{ article.title }}</span>
                <span class="news-source">{{ article.source }}</span>
              </div>
            </a>
          </div>
        </div>

        <!-- Empty state -->
        <div
          class="empty-state"
          *ngIf="!weather && flights.length === 0 && maritimeData.length === 0 && news.length === 0"
        >
          <p>Click anywhere on the map to load data for that location.</p>
        </div>
      </div>
    </div>
  `,
  styles: [
    `
      .data-panel {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        background: #1a1a2e;
        border-top: 1px solid #2a2a4a;
        max-height: 60px;
        transition: max-height 0.3s ease;
        overflow: hidden;
        z-index: 1000;
      }

      .data-panel.expanded {
        max-height: 45vh;
      }

      .panel-handle {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 8px;
        cursor: pointer;
        gap: 10px;
      }

      .handle-bar {
        width: 40px;
        height: 4px;
        background: #3a3a5a;
        border-radius: 2px;
      }

      .panel-title {
        color: #8888aa;
        font-size: 12px;
        text-transform: uppercase;
        letter-spacing: 1px;
      }

      .panel-content {
        display: flex;
        gap: 16px;
        padding: 0 16px 16px;
        overflow-x: auto;
        overflow-y: hidden;
      }

      .data-card {
        min-width: 260px;
        max-width: 320px;
        background: #0f0f23;
        border-radius: 10px;
        padding: 14px;
        border: 1px solid #2a2a4a;
        flex-shrink: 0;
      }

      .card-header {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 12px;
      }

      .card-icon {
        font-size: 18px;
      }

      .card-header h4 {
        margin: 0;
        color: #e0e0ff;
        font-size: 14px;
        flex: 1;
      }

      .badge {
        background: #4a5aff;
        color: #fff;
        font-size: 11px;
        padding: 2px 8px;
        border-radius: 10px;
        font-weight: 600;
      }

      .weather-info {
        display: flex;
        gap: 16px;
        align-items: center;
      }

      .temp {
        font-size: 32px;
        color: #fff;
        font-weight: 300;
      }

      .weather-details {
        display: flex;
        flex-direction: column;
        gap: 2px;
        color: #8888aa;
        font-size: 12px;
      }

      .flight-list,
      .marina-list {
        display: flex;
        flex-direction: column;
        gap: 4px;
      }

      .flight-item,
      .marina-item {
        display: flex;
        justify-content: space-between;
        padding: 4px 8px;
        border-radius: 4px;
        background: rgba(255, 255, 255, 0.02);
        font-size: 12px;
        color: #aaa;
      }

      .callsign,
      .marina-name {
        color: #e0e0ff;
        font-weight: 500;
      }

      .more-indicator {
        text-align: center;
        color: #6678ff;
        font-size: 12px;
        padding: 4px;
      }

      .news-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }

      .news-item {
        display: flex;
        gap: 10px;
        text-decoration: none;
        padding: 6px;
        border-radius: 6px;
        transition: background 0.2s;
      }

      .news-item:hover {
        background: rgba(255, 255, 255, 0.05);
      }

      .news-thumb {
        width: 48px;
        height: 48px;
        border-radius: 6px;
        object-fit: cover;
        flex-shrink: 0;
      }

      .news-text {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
      }

      .news-title {
        color: #ccc;
        font-size: 12px;
        line-height: 1.3;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }

      .news-source {
        color: #6678ff;
        font-size: 11px;
      }

      .empty-state {
        color: #666;
        font-size: 13px;
        padding: 20px;
        text-align: center;
        width: 100%;
      }
    `,
  ],
})
export class DataPanelComponent {
  @Input() weather: WeatherData | null = null;
  @Input() flights: FlightData[] = [];
  @Input() maritimeData: MaritimeData[] = [];
  @Input() news: NewsArticle[] = [];

  expanded = true;
}
