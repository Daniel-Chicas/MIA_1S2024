import React, { Component } from "react";
import { MenuItem, Menu } from "semantic-ui-react";
import "./Navs.css";

export default class MenuExampleBasic extends Component {
    render() {
        return (
            <Menu inverted className="Nav">
                <MenuItem href="/">Pantalla 1</MenuItem>
                <MenuItem href="/vista2">Pantalla 2</MenuItem>
                <MenuItem href="/vista3">Pantalla 3</MenuItem>
            </Menu>
        );
    }
}
