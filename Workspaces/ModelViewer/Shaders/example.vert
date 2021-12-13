#version 330
in vec3 vert;
in vec3 normal;


uniform mat4 modelMatrix;
uniform mat4 viewMatrix;
uniform mat4 projMatrix;
uniform mat4 MVP;


out vec2 fragTexCoord;
out vec3 fragNormal;
out vec3 fragVert;
out vec3 fragWorldPos;

void main() {
    //fragTexCoord = vertTexCoord;
    fragNormal = normal;
    fragVert = vert;
    fragWorldPos = (modelMatrix * vec4(vert,1)).xyz;
	gl_Position = MVP * vec4(vert, 1);
}