import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  constructor(
    private httpClient: HttpClient
  ) { }

  postEntrada(entrada: string) {
    return this.httpClient.post("http://54.164.50.60:5000/analizar", { Cmd: entrada });
    //return this.httpClient.post("http://XX.XX.XXX.XXX:5000/analizar", { Cmd: entrada });
  }
}
