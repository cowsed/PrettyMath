#version 330
in vec3 vp;

in vec2 uv_in;
out vec2 UV;

void main() {
	UV=vp.yx;
	gl_Position = vec4(vp, 1.0);
}