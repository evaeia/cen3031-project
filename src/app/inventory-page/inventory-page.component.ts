import { Component, OnInit } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { DialogComponent } from './dialog/dialog.component';
import { ConfirmDialogComponent } from './confirm-dialog/confirm-dialog.component';

@Component({
  selector: 'app-inventory-page',
  templateUrl: './inventory-page.component.html',
  styleUrls: ['./inventory-page.component.css']
})
export class InventoryPageComponent implements OnInit {
  containers = [ 
    {
      id: 1,
      name: 'Fridge',
      description: "All my food is kept in here."
    }, 
    { 
      id: 2,
      name: 'Workbench',
      description: "All my tools are kept in here."
    }, 
    {
      id: 3,
      name: 'Dresser',
      description: "All my clothes are kept in here."
    }];

    constructor(public dialog: MatDialog) {}

    ngOnInit() {}

    openDialog(): void {
      const dialogRef = this.dialog.open(DialogComponent, {
        data: {name: '', description: ''}
      });
  
      dialogRef.afterClosed().subscribe(result => {
        if (result) {
          this.containers.push({id: this.containers[this.containers.length - 1].id + 1, name: result.name, description: result.description});
        }
      });
    }

    removeContainer(index: number) {
      this.containers.splice(index, 1);
    }

    openConfirmDialog(index: number) {
      const dialogRef = this.dialog.open(ConfirmDialogComponent, {
        width: '250px',
        data: { name: this.containers[index].name }
      });
  
      dialogRef.afterClosed().subscribe(result => {
        if (result) {
          this.removeContainer(index);
        }
      });
    }
}
