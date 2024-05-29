import "./site.css";

function setup() {
  let templates = document.querySelectorAll('template[simple]');
  templates.forEach((template) => {
    let templateId = template.getAttribute('id');
    let templateContent = template.content;

    customElements.define(templateId, class extends HTMLElement {
      constructor() {
        super();
        this.appendChild(templateContent.cloneNode(true));
        this.slots = this.querySelectorAll('slot');
      }

      connectedCallback() {
        let toremove = [];
        this.slots.forEach((tslot) => {
          let slotName = tslot.getAttribute('name');
          let slotContent = this.querySelector(`[slot="${slotName}"]`);
          if (slotContent) {
            tslot.replaceWith(slotContent.cloneNode(true));
            toremove.push(slotContent);
          }
        });
        toremove.forEach((element) => {
            element.remove();
        });
      }
    }); 
  });
}

export { setup } 
