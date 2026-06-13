import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map, catchError, of } from 'rxjs';
import { FlightData } from '../models/data.models';

@Injectable({ providedIn: 'root' })
export class FlightService {
  private readonly apiUrl = 'https://opensky-network.org/api';

  constructor(private http: HttpClient) {}

  getFlightsInBounds(
    lamin: number,
    lomin: number,
    lamax: number,
    lomax: number
  ): Observable<FlightData[]> {
    const url = `${this.apiUrl}/states/all?lamin=${lamin}&lomin=${lomin}&lamax=${lamax}&lomax=${lomax}`;

    return this.http.get<any>(url).pipe(
      map((response) => {
        if (!response?.states) return [];
        return response.states
          .filter((s: any[]) => s[5] != null && s[6] != null)
          .map((s: any[]) => ({
            icao24: s[0],
            callsign: (s[1] || '').trim(),
            originCountry: s[2],
            longitude: s[5],
            latitude: s[6],
            altitude: s[7] || 0,
            velocity: s[9] || 0,
            heading: s[10] || 0,
            onGround: s[8],
          }));
      }),
      catchError((err) => {
        console.error('Flight API error:', err);
        return of([]);
      })
    );
  }
}
