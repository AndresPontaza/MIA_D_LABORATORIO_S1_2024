import { Component, OnInit } from '@angular/core';
import { graphviz } from "d3-graphviz"
import { ApiService } from 'src/app/services/api.service';

@Component({
  selector: 'app-reportes',
  templateUrl: './reportes.component.html',
  styleUrls: ['./reportes.component.css']
})
export class ReportesComponent implements OnInit {
  entrada = "";

  constructor(public service: ApiService) { }

  ngOnInit(): void {
  }

  ejecutar() {
    if (this.entrada != "") {
      this.service.postEntrada(this.entrada).subscribe(async (res: any) => {
        graphviz("#graph").renderDot(String(res.result))
      });
    } else {
      alert("Entrada vacia...")
    }
  }
}
