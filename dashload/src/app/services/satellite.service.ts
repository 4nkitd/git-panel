import { Injectable } from '@angular/core';
import { SatelliteLayer } from '../models/data.models';

@Injectable({ providedIn: 'root' })
export class SatelliteService {
  private readonly nasaGibsBase =
    'https://gibs.earthdata.nasa.gov/wmts/epsg3857/best';

  getLayers(): SatelliteLayer[] {
    return [
      {
        id: 'modis-terra',
        name: 'MODIS Terra (True Color)',
        tileUrl: `${this.nasaGibsBase}/MODIS_Terra_CorrectedReflectance_TrueColor/default/{time}/GoogleMapsCompatible_Level9/{z}/{y}/{x}.jpg`,
        attribution: 'NASA EOSDIS GIBS',
        opacity: 0.8,
        enabled: false,
      },
      {
        id: 'viirs-snpp',
        name: 'VIIRS Night Lights',
        tileUrl: `${this.nasaGibsBase}/VIIRS_SNPP_DayNightBand_AtSensor_M15/default/{time}/GoogleMapsCompatible_Level8/{z}/{y}/{x}.png`,
        attribution: 'NASA EOSDIS GIBS',
        opacity: 0.7,
        enabled: false,
      },
      {
        id: 'modis-aqua',
        name: 'MODIS Aqua (True Color)',
        tileUrl: `${this.nasaGibsBase}/MODIS_Aqua_CorrectedReflectance_TrueColor/default/{time}/GoogleMapsCompatible_Level9/{z}/{y}/{x}.jpg`,
        attribution: 'NASA EOSDIS GIBS',
        opacity: 0.8,
        enabled: false,
      },
      {
        id: 'sea-surface-temp',
        name: 'Sea Surface Temperature',
        tileUrl: `${this.nasaGibsBase}/GHRSST_L4_MUR_Sea_Surface_Temperature/default/{time}/GoogleMapsCompatible_Level7/{z}/{y}/{x}.png`,
        attribution: 'NASA EOSDIS GIBS',
        opacity: 0.6,
        enabled: false,
      },
    ];
  }

  getTileUrlForDate(layer: SatelliteLayer, date?: Date): string {
    const d = date || new Date(Date.now() - 86400000); // yesterday (today may not be available)
    const dateStr = d.toISOString().split('T')[0];
    return layer.tileUrl.replace('{time}', dateStr);
  }
}
