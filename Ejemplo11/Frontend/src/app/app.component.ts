import { Component } from '@angular/core';
import { ApiService } from './services/api.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'ejemplo11-web';

  constructor(public service: ApiService) { }

  comando = "";

  Salir() {
    this.comando = "logout";
    this.service.postEntrada(this.comando).subscribe(async (res: any) => {
      alert(await res.result + "\n");
    });
  }
}
