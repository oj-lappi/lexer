/*
This parser is driven by metadata.

It is a recursive descent parser where every Non-terminal nodetype
has a production-function assigned in a map.

The map is populated by calling a second-order function closured
with the metadata describing the production.
*/
package Gparse
