export interface MapPosition {
  lat: number;
  lng: number;
  zoom: number;
}

export interface FlightData {
  icao24: string;
  callsign: string;
  originCountry: string;
  longitude: number;
  latitude: number;
  altitude: number;
  velocity: number;
  heading: number;
  onGround: boolean;
}

export interface MaritimeData {
  id: number;
  name: string;
  lat: number;
  lng: number;
  type: string;
  tags: Record<string, string>;
}

export interface WeatherData {
  lat: number;
  lng: number;
  temperature: number;
  windSpeed: number;
  windDirection: number;
  weatherCode: number;
  humidity: number;
  description: string;
  icon: string;
}

export interface NewsArticle {
  title: string;
  url: string;
  source: string;
  publishedAt: string;
  imageUrl?: string;
  summary?: string;
  location?: string;
  category?: string;
}

export interface SatelliteLayer {
  id: string;
  name: string;
  tileUrl: string;
  attribution: string;
  opacity: number;
  enabled: boolean;
}

export type DataCategory = 'flights' | 'maritime' | 'satellite' | 'weather' | 'news';

export interface LayerToggle {
  category: DataCategory;
  label: string;
  icon: string;
  enabled: boolean;
  count?: number;
}
