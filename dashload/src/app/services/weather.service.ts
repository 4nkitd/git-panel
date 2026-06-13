import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map, catchError, of } from 'rxjs';
import { WeatherData } from '../models/data.models';

const WMO_CODES: Record<number, { description: string; icon: string }> = {
  0: { description: 'Clear sky', icon: '☀️' },
  1: { description: 'Mainly clear', icon: '🌤️' },
  2: { description: 'Partly cloudy', icon: '⛅' },
  3: { description: 'Overcast', icon: '☁️' },
  45: { description: 'Fog', icon: '🌫️' },
  48: { description: 'Rime fog', icon: '🌫️' },
  51: { description: 'Light drizzle', icon: '🌦️' },
  53: { description: 'Moderate drizzle', icon: '🌦️' },
  55: { description: 'Dense drizzle', icon: '🌧️' },
  61: { description: 'Slight rain', icon: '🌧️' },
  63: { description: 'Moderate rain', icon: '🌧️' },
  65: { description: 'Heavy rain', icon: '🌧️' },
  71: { description: 'Slight snow', icon: '🌨️' },
  73: { description: 'Moderate snow', icon: '🌨️' },
  75: { description: 'Heavy snow', icon: '❄️' },
  80: { description: 'Slight showers', icon: '🌦️' },
  81: { description: 'Moderate showers', icon: '🌧️' },
  82: { description: 'Violent showers', icon: '⛈️' },
  95: { description: 'Thunderstorm', icon: '⛈️' },
  96: { description: 'Thunderstorm w/ hail', icon: '⛈️' },
  99: { description: 'Thunderstorm w/ heavy hail', icon: '⛈️' },
};

@Injectable({ providedIn: 'root' })
export class WeatherService {
  private readonly apiUrl = 'https://api.open-meteo.com/v1/forecast';

  constructor(private http: HttpClient) {}

  getWeather(lat: number, lng: number): Observable<WeatherData> {
    const url = `${this.apiUrl}?latitude=${lat}&longitude=${lng}&current=temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m,wind_direction_10m`;

    return this.http.get<any>(url).pipe(
      map((response) => {
        const current = response.current;
        const code = current.weather_code;
        const wmo = WMO_CODES[code] || {
          description: 'Unknown',
          icon: '❓',
        };

        return {
          lat,
          lng,
          temperature: current.temperature_2m,
          windSpeed: current.wind_speed_10m,
          windDirection: current.wind_direction_10m,
          weatherCode: code,
          humidity: current.relative_humidity_2m,
          description: wmo.description,
          icon: wmo.icon,
        };
      }),
      catchError((err) => {
        console.error('Weather API error:', err);
        return of({
          lat,
          lng,
          temperature: 0,
          windSpeed: 0,
          windDirection: 0,
          weatherCode: -1,
          humidity: 0,
          description: 'Unavailable',
          icon: '❓',
        });
      })
    );
  }
}
