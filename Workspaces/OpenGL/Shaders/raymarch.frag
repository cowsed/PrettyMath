#version 330
in vec2 UV;
layout(location = 0) out vec4 frag_color;

//Uniforms
uniform vec3 clearColor;
uniform vec3 camPos;
uniform vec3 lookAt;
//Rendering things
uniform int MAX_MARCHING_STEPS;

//Control over scene
uniform float iTime;
uniform int DBDRAW;

//Constants
const float MIN_DIST = 0.0;
const float MAX_DIST = 100.0;
const float EPSILON = 0.0001;

struct sdResult{
    float distance;
    int materialID;
};


sdResult sdMin(sdResult a, sdResult b){
    if (a.distance<b.distance){
        return a;
    }
    return b;
}

sdResult sdSphere(vec3 samplePoint, float radius, int matID) {
    sdResult res;
    res.distance = length(samplePoint) - radius;
    res.materialID = matID;
    return res;
}

sdResult sdBox( vec3 p, vec3 b, int matID)
{
  vec3 q = abs(p) - b;
  sdResult res;
  res.distance = length(max(q,0.0)) + min(max(q.x,max(q.y,q.z)),0.0);
  res.materialID = matID;
  return res;
}

sdResult sdXYPlane(vec3 p, float z, int matID){
    sdResult res;
    res.distance = p.z-z;
    res.materialID = matID;
    return res;
}

sdResult sceneSDF(vec3 samplePoint) {
    sdResult res;
    res = sdMin(  
        sdBox(samplePoint, vec3(1,1,1), 0), 
        sdMin(
            sdSphere(samplePoint-vec3(0,0,1.25),.5,1),
            sdXYPlane(samplePoint,-2,2))
            );
    return res;
}

struct DistRes {
   float dist;
   int steps;
};

DistRes shortestDistanceToSurface(vec3 eye, vec3 marchingDirection, float start, float end) {
    DistRes result;
    float depth = start;
    int i;
    for (i = 0; i < MAX_MARCHING_STEPS; i++) {
        sdResult res =sceneSDF(eye + depth * marchingDirection); 
        float dist = res.distance;
        if (dist < EPSILON) {
			result.dist = depth;
            result.steps = i;
			return result;
        }
        depth += dist;
        if (depth >= end) {
            result.dist = end;
            result.steps = i;
            return result;
        }
    }
    result.dist = end;        
    result.steps = i;
    return result;
    
}





vec3 estimateNormal(vec3 p) {
    return normalize(vec3(
        sceneSDF(vec3(p.x + EPSILON, p.y, p.z)).distance - sceneSDF(vec3(p.x - EPSILON, p.y, p.z)).distance,
        sceneSDF(vec3(p.x, p.y + EPSILON, p.z)).distance - sceneSDF(vec3(p.x, p.y - EPSILON, p.z)).distance,
        sceneSDF(vec3(p.x, p.y, p.z  + EPSILON)).distance - sceneSDF(vec3(p.x, p.y, p.z - EPSILON)).distance
    ));
}
/**
 * Lighting contribution of a single point light source via Phong illumination.
 * 
 * The vec3 returned is the RGB color of the light's contribution.
 *
 * k_a: Ambient color
 * k_d: Diffuse color
 * k_s: Specular color
 * alpha: Shininess coefficient
 * p: position of point being lit
 * eye: the position of the camera
 * lightPos: the position of the light
 * lightIntensity: color/intensity of the light
 *
 * See https://en.wikipedia.org/wiki/Phong_reflection_model#Description
 */
vec3 phongContribForLight(vec3 k_d, vec3 k_s, float alpha, vec3 p, vec3 eye,
                          vec3 lightPos, vec3 lightIntensity) {
    vec3 N = estimateNormal(p);
    vec3 L = normalize(lightPos - p);
    vec3 V = normalize(eye - p);
    vec3 R = normalize(reflect(-L, N));
    
    float dotLN = dot(L, N);
    float dotRV = dot(R, V);
    
    if (dotLN < 0.0) {
        // Light not visible from this point on the surface
        return vec3(0.0, 0.0, 0.0);
    } 
    
    if (dotRV < 0.0) {
        // Light reflection in opposite direction as viewer, apply only diffuse
        // component
        return lightIntensity * (k_d * dotLN);
    }
    return lightIntensity * (k_d * dotLN + k_s * pow(dotRV, alpha));
}


/**
 * Lighting via Phong illumination.
 * 
 * The vec3 returned is the RGB color of that point after lighting is applied.
 * k_a: Ambient color
 * k_d: Diffuse color
 * k_s: Specular color
 * alpha: Shininess coefficient
 * p: position of point being lit
 * eye: the position of the camera
 *
 * See https://en.wikipedia.org/wiki/Phong_reflection_model#Description
 */
vec3 phongIllumination(vec3 k_a, vec3 k_d, vec3 k_s, float alpha, vec3 p, vec3 eye) {
    const vec3 ambientLight = 0.5 * vec3(1.0, 1.0, 1.0);
    vec3 color = ambientLight * k_a;
    
    vec3 light1Pos = vec3(4.0 * sin(iTime),
                          2.0,
                          4.0 * cos(iTime));
    vec3 light1Intensity = vec3(0.4, 0.4, 0.4);
    
    color += phongContribForLight(k_d, k_s, alpha, p, eye,
                                  light1Pos,
                                  light1Intensity);
    
    return color;
}




//uv should be x[-1,1] and y[-1,1]

vec3 rayDirection(float fieldOfView, vec2 uv) {
    vec2 xy = uv;
    float z = 1 / tan(radians(fieldOfView) / 2.0);
    return normalize(vec3(xy, -z));
}


mat4 viewMatrix(vec3 eye, vec3 center, vec3 up) {
	vec3 f = normalize(center - eye);
	vec3 s = normalize(cross(f, up));
	vec3 u = cross(s, f);
	return mat4(
		vec4(s, 0.0),
		vec4(u, 0.0),
		vec4(-f, 0.0),
		vec4(0.0, 0.0, 0.0, 1)
	);
}

void main() {
    float aspect = 8.0/6.0;
    vec2 uv = UV/vec2(aspect, 1);

    
	vec3 dir = rayDirection(45.0, uv);
    vec3 eye = camPos;
    
        
    mat4 viewToWorld = viewMatrix(eye, lookAt, vec3(0.0, 1.0, 0.0));
    
    vec3 worldDir = (viewToWorld * vec4(dir, 0.0)).xyz;
    
    
    DistRes Res = shortestDistanceToSurface(eye, worldDir, MIN_DIST, MAX_DIST);
    
    float dist = Res.dist;
    
    if (dist > MAX_DIST - EPSILON) {
        // Didn't hit anything
        if (DBDRAW!=0)
            frag_color = vec4(clearColor, 1.0);
		return;
    }


    vec3 colors[3];
    colors[0] = vec3(1,0,0);
    colors[1] = vec3(1,0,1);

    // The closest point on the surface to the eyepoint along the view ray
    vec3 p = eye + dist * worldDir;

    bool map = floor(mod(p.x,2)) == 0 ^^ floor(mod(p.y,2)) == 0;

    colors[2] = mix(vec3(1,1,1), vec3(0,0,0), float(map));

            
    int matID = sceneSDF(p).materialID;
   
    vec3 K_a = clearColor;//vec3(0.2, 0.2, 0.2);
    vec3 K_d = colors[matID];
    vec3 K_s = vec3(1.0, 1.0, 1.0);
    float shininess = 20.0;
    
    vec3 color = phongIllumination(K_a, K_d, K_s,shininess, p, eye);

    if (DBDRAW == 0){
        color = vec3(float(Res.steps)/float(MAX_MARCHING_STEPS),0,0);
    }
    frag_color = vec4(color, 1.0);


}




