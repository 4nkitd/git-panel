import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map, catchError, of } from 'rxjs';
import { MaritimeData } from '../models/data.models';

@Injectable({ providedIn: 'root' })
export class MaritimeService {
  private readonly overpassUrl = 'https://overpass-api.de/api/interpreter';

  constructor(private http: HttpClient) {}

  getMarinasInBounds(
    south: number,
    west: number,
    north: number,
    east: number
  ): Observable<MaritimeData[]> {
    const query = `
      [out:json][timeout:15];
      (
        node["leisure"="marina"](${south},${west},${north},${east});
        way["leisure"="marina"](${south},${west},${north},${east});
        node["seamark:type"](${south},${west},${north},${east});
        node["harbour"="yes"](${south},${west},${north},${east});
      );
      out center body;
    `;

    return this.http
      .post(this.overpassUrl, `data=${encodeURIComponent(query)}`, {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      })
      .pipe(
        map((response: any) => {
          if (!response?.elements) return [];
          return response.elements.map((el: any) => ({
            id: el.id,
            name: el.tags?.name || el.tags?.['seamark:name'] || 'Unknown',
            lat: el.lat || el.center?.lat,
            lng: el.lon || el.center?.lon,
            type: el.tags?.['seamark:type'] || el.tags?.leisure || 'marina',
            tags: el.tags || {},
          }));
        }),
        catchError((err) => {
          console.error('Maritime API error:', err);
          return of([]);
        })
      );
  }
}
