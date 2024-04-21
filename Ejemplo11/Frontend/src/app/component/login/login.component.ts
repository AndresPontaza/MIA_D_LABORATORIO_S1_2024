import { Component, OnInit } from '@angular/core';
import { NgForm } from '@angular/forms';
import { ApiService } from 'src/app/services/api.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {

  comando = "";

  constructor(public service: ApiService) { }

  ngOnInit(): void {
  }

  ingresar(form: NgForm) {
    this.comando = "login -usuario=\"" + form.value.user + "\" -password=\"" + form.value.pass + "\" -id=" + form.value.id_particion;
    this.service.postEntrada(this.comando).subscribe(async (res: any) => {
      alert(await res.result + "\n");
    });
  }

}
