import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  DataCategory,
  LayerToggle,
  SatelliteLayer,
} from '../../models/data.models';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <aside class="sidebar" [class.collapsed]="collapsed">
      <button class="collapse-btn" (click)="collapsed = !collapsed">
        {{ collapsed ? '→' : '←' }}
      </button>

      <div class="sidebar-content" *ngIf="!collapsed">
        <div class="sidebar-header">
          <h2>Layers</h2>
        </div>

        <div class="layer-section">
          <div
            class="layer-toggle"
            *ngFor="let layer of layers"
            [class.active]="layer.enabled"
            (click)="toggleLayer(layer)"
          >
            <span class="layer-icon">{{ layer.icon }}</span>
            <span class="layer-label">{{ layer.label }}</span>
            <span class="layer-count" *ngIf="layer.count !== undefined">{{
              layer.count
            }}</span>
            <span class="layer-switch" [class.on]="layer.enabled"></span>
          </div>
        </div>

        <div class="section-divider"></div>

        <div class="satellite-section">
          <h3>Satellite Imagery</h3>
          <div
            class="satellite-toggle"
            *ngFor="let sat of satelliteLayers"
            [class.active]="sat.enabled"
            (click)="toggleSatellite(sat)"
          >
            <span class="sat-name">{{ sat.name }}</span>
            <span class="layer-switch" [class.on]="sat.enabled"></span>
          </div>
        </div>

        <div class="section-divider"></div>

        <div class="news-category-section">
          <h3>News Category</h3>
          <select
            class="category-select"
            [(ngModel)]="selectedCategory"
            (ngModelChange)="categoryChanged.emit($event)"
          >
            <option value="">All News</option>
            <option value="aviation">Aviation</option>
            <option value="maritime">Maritime</option>
            <option value="weather">Weather</option>
            <option value="disaster">Disasters</option>
            <option value="politics">Politics</option>
            <option value="technology">Technology</option>
          </select>
        </div>

        <div class="section-divider"></div>

        <div class="location-section">
          <h3>Quick Locations</h3>
          <button
            class="location-btn"
            *ngFor="let loc of quickLocations"
            (click)="locationSelected.emit(loc)"
          >
            {{ loc.name }}
          </button>
        </div>
      </div>
    </aside>
  `,
  styles: [
    `
      .sidebar {
        width: 280px;
        height: 100%;
        background: #1a1a2e;
        border-right: 1px solid #2a2a4a;
        display: flex;
        flex-direction: column;
        transition: width 0.3s ease;
        position: relative;
        overflow: hidden;
      }

      .sidebar.collapsed {
        width: 40px;
      }

      .collapse-btn {
        position: absolute;
        top: 10px;
        right: 8px;
        background: #2a2a4a;
        border: none;
        color: #8888aa;
        width: 24px;
        height: 24px;
        border-radius: 4px;
        cursor: pointer;
        z-index: 2;
        font-size: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .collapse-btn:hover {
        background: #3a3a5a;
        color: #fff;
      }

      .sidebar-content {
        padding: 48px 16px 16px;
        overflow-y: auto;
        flex: 1;
      }

      .sidebar-header h2 {
        color: #e0e0ff;
        font-size: 16px;
        margin: 0 0 16px;
        font-weight: 600;
      }

      h3 {
        color: #9999bb;
        font-size: 12px;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin: 0 0 10px;
      }

      .layer-toggle,
      .satellite-toggle {
        display: flex;
        align-items: center;
        padding: 10px 12px;
        border-radius: 8px;
        cursor: pointer;
        margin-bottom: 4px;
        transition: background 0.2s;
      }

      .layer-toggle:hover,
      .satellite-toggle:hover {
        background: rgba(255, 255, 255, 0.05);
      }

      .layer-toggle.active,
      .satellite-toggle.active {
        background: rgba(100, 120, 255, 0.1);
      }

      .layer-icon {
        font-size: 18px;
        margin-right: 10px;
        width: 24px;
        text-align: center;
      }

      .layer-label,
      .sat-name {
        flex: 1;
        color: #ccc;
        font-size: 14px;
      }

      .layer-count {
        color: #6678ff;
        font-size: 12px;
        margin-right: 8px;
        font-weight: 600;
      }

      .layer-switch {
        width: 36px;
        height: 20px;
        background: #333;
        border-radius: 10px;
        position: relative;
        transition: background 0.2s;
      }

      .layer-switch::after {
        content: '';
        position: absolute;
        width: 16px;
        height: 16px;
        background: #666;
        border-radius: 50%;
        top: 2px;
        left: 2px;
        transition: all 0.2s;
      }

      .layer-switch.on {
        background: #4a5aff;
      }

      .layer-switch.on::after {
        background: #fff;
        left: 18px;
      }

      .section-divider {
        height: 1px;
        background: #2a2a4a;
        margin: 16px 0;
      }

      .category-select {
        width: 100%;
        padding: 8px 12px;
        background: #0f0f23;
        border: 1px solid #2a2a4a;
        border-radius: 6px;
        color: #ccc;
        font-size: 13px;
        outline: none;
      }

      .category-select:focus {
        border-color: #4a5aff;
      }

      .location-btn {
        display: block;
        width: 100%;
        padding: 8px 12px;
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid #2a2a4a;
        border-radius: 6px;
        color: #aaa;
        font-size: 13px;
        cursor: pointer;
        margin-bottom: 6px;
        text-align: left;
        transition: all 0.2s;
      }

      .location-btn:hover {
        background: rgba(100, 120, 255, 0.1);
        border-color: #4a5aff;
        color: #fff;
      }
    `,
  ],
})
export class SidebarComponent {
  @Input() layers: LayerToggle[] = [];
  @Input() satelliteLayers: SatelliteLayer[] = [];

  @Output() layerToggled = new EventEmitter<DataCategory>();
  @Output() satelliteToggled = new EventEmitter<SatelliteLayer>();
  @Output() categoryChanged = new EventEmitter<string>();
  @Output() locationSelected = new EventEmitter<{
    name: string;
    lat: number;
    lng: number;
    zoom: number;
  }>();

  collapsed = false;
  selectedCategory = '';

  quickLocations = [
    { name: 'New York', lat: 40.7128, lng: -74.006, zoom: 10 },
    { name: 'London', lat: 51.5074, lng: -0.1278, zoom: 10 },
    { name: 'Tokyo', lat: 35.6762, lng: 139.6503, zoom: 10 },
    { name: 'Dubai', lat: 25.2048, lng: 55.2708, zoom: 10 },
    { name: 'Sydney', lat: -33.8688, lng: 151.2093, zoom: 10 },
    { name: 'Mumbai', lat: 19.076, lng: 72.8777, zoom: 11 },
  ];

  toggleLayer(layer: LayerToggle): void {
    layer.enabled = !layer.enabled;
    this.layerToggled.emit(layer.category);
  }

  toggleSatellite(layer: SatelliteLayer): void {
    layer.enabled = !layer.enabled;
    this.satelliteToggled.emit(layer);
  }
}
