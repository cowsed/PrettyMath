#version 330
in vec2 UV;

layout(location = 0) out vec4 frag_colour;
void main() {
    vec2 uv = UV;
    vec3 col = vec3(uv,0);
    frag_colour = vec4(col, 1);
}