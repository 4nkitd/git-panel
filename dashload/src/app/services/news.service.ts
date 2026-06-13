import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map, catchError, of } from 'rxjs';
import { NewsArticle } from '../models/data.models';

@Injectable({ providedIn: 'root' })
export class NewsService {
  private readonly gdeltUrl = 'https://api.gdeltproject.org/api/v2/doc/doc';

  constructor(private http: HttpClient) {}

  getNewsByLocation(
    lat: number,
    lng: number,
    category?: string
  ): Observable<NewsArticle[]> {
    const nearQuery = `near:${lat},${lng},50km`;
    const themeQuery = category ? ` ${this.mapCategoryToTheme(category)}` : '';
    const query = encodeURIComponent(nearQuery + themeQuery);

    const url = `${this.gdeltUrl}?query=${query}&mode=artlist&maxrecords=20&format=json&sort=datedesc`;

    return this.http.get<any>(url).pipe(
      map((response) => {
        if (!response?.articles) return [];
        return response.articles.map((article: any) => ({
          title: article.title,
          url: article.url,
          source: article.domain || article.source || 'Unknown',
          publishedAt: article.seendate,
          imageUrl: article.socialimage || undefined,
          summary: article.title,
          location: `${lat.toFixed(2)}, ${lng.toFixed(2)}`,
          category: category || 'general',
        }));
      }),
      catchError((err) => {
        console.error('News API error:', err);
        return of([]);
      })
    );
  }

  getNewsByKeyword(keyword: string): Observable<NewsArticle[]> {
    const query = encodeURIComponent(keyword);
    const url = `${this.gdeltUrl}?query=${query}&mode=artlist&maxrecords=20&format=json&sort=datedesc`;

    return this.http.get<any>(url).pipe(
      map((response) => {
        if (!response?.articles) return [];
        return response.articles.map((article: any) => ({
          title: article.title,
          url: article.url,
          source: article.domain || 'Unknown',
          publishedAt: article.seendate,
          imageUrl: article.socialimage || undefined,
          summary: article.title,
          category: 'general',
        }));
      }),
      catchError(() => of([]))
    );
  }

  private mapCategoryToTheme(category: string): string {
    const themes: Record<string, string> = {
      aviation: 'airline OR aviation OR flight',
      maritime: 'maritime OR shipping OR port',
      weather: 'weather OR storm OR climate',
      disaster: 'earthquake OR hurricane OR flood',
      politics: 'politics OR government OR election',
      technology: 'technology OR AI OR software',
    };
    return themes[category] || category;
  }
}
