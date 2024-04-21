import { Component, OnInit } from '@angular/core';
import { ApiService } from 'src/app/services/api.service';

@Component({
  selector: 'app-inicio',
  templateUrl: './inicio.component.html',
  styleUrls: ['./inicio.component.css']
})
export class InicioComponent implements OnInit {

  entrada = "";
  salida = "";

  constructor(public service: ApiService) { }

  ngOnInit(): void { }

  public async onFileSelected(event: any) {
    const file: File = event.target.files[0];
    this.entrada = await file.text();
  }

  ejecutar() {
    this.salida = "Enviando comandos al API por favor espere..."
    if (this.entrada != "") {
      this.service.postEntrada(this.entrada).subscribe(async (res: any) => {
        this.salida = res.result;
      });
    } else {
      alert("Archivo de entrada vacio...")
      this.salida = "";
    }
  }
}