import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <header class="header">
      <div class="brand">
        <span class="logo">◈</span>
        <h1>Dashload</h1>
      </div>

      <div class="search-bar">
        <input
          type="text"
          [(ngModel)]="searchQuery"
          placeholder="Search location or news..."
          (keydown.enter)="onSearch()"
        />
        <button class="search-btn" (click)="onSearch()">Search</button>
      </div>

      <div class="header-info">
        <div class="coord-display" *ngIf="currentLat !== null">
          <span>{{ currentLat | number:'1.4-4' }}, {{ currentLng | number:'1.4-4' }}</span>
        </div>
        <div class="status-dot" [class.loading]="loading"></div>
      </div>
    </header>
  `,
  styles: [
    `
      .header {
        display: flex;
        align-items: center;
        padding: 0 20px;
        height: 56px;
        background: #0f0f23;
        border-bottom: 1px solid #2a2a4a;
        gap: 20px;
      }

      .brand {
        display: flex;
        align-items: center;
        gap: 10px;
        flex-shrink: 0;
      }

      .logo {
        font-size: 24px;
        color: #4a5aff;
      }

      h1 {
        margin: 0;
        font-size: 20px;
        font-weight: 700;
        color: #e0e0ff;
        letter-spacing: -0.5px;
      }

      .search-bar {
        flex: 1;
        max-width: 480px;
        display: flex;
        gap: 8px;
      }

      input {
        flex: 1;
        padding: 8px 14px;
        background: #1a1a2e;
        border: 1px solid #2a2a4a;
        border-radius: 8px;
        color: #ccc;
        font-size: 13px;
        outline: none;
      }

      input:focus {
        border-color: #4a5aff;
      }

      input::placeholder {
        color: #555;
      }

      .search-btn {
        padding: 8px 16px;
        background: #4a5aff;
        border: none;
        border-radius: 8px;
        color: #fff;
        font-size: 13px;
        cursor: pointer;
        font-weight: 500;
        transition: background 0.2s;
      }

      .search-btn:hover {
        background: #5a6aff;
      }

      .header-info {
        display: flex;
        align-items: center;
        gap: 12px;
        flex-shrink: 0;
      }

      .coord-display {
        color: #6678ff;
        font-size: 12px;
        font-family: monospace;
      }

      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #4caf50;
      }

      .status-dot.loading {
        background: #ff9800;
        animation: pulse 1s infinite;
      }

      @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.4; }
      }
    `,
  ],
})
export class HeaderComponent {
  @Input() loading = false;
  @Input() currentLat: number | null = null;
  @Input() currentLng: number | null = null;

  @Output() search = new EventEmitter<string>();

  searchQuery = '';

  onSearch(): void {
    if (this.searchQuery.trim()) {
      this.search.emit(this.searchQuery.trim());
    }
  }
}
